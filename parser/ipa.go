package parser

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/poolqa/CgbiPngFix/ipaPng"
	"howett.net/plist"
	"icu.bughub.app/ipc-tool/model"
)

// - 解压 ipa 包后得到 xxx.app
//
// - 执行以下命令可将 xxx.app 得到签名文件
//
// ```
// codesign -d --extract-certificates xxxx.app
// ```
//
// - 执行以下命令获取得 sha-1 和 modulus 信息
//
// ```
// openssl x509 -fingerprint -sha1 -modulus -text -noout -in codesign0
// ```
//
// - 解析 xxx.app 中的 Info.plist 文件得到 bundle id 和 名称以及图标路径
//
// - 使用以下合集还原图片以保证在其他终端正常显示
//
// ```
// xcrun -sdk iphoneos pngcrush \ -q -revert-iphone-optimizations -d AppIcon60x60@2x.png
// ```
func (app *App) ParseIpa(path string) (*model.Feature, error) {
	//  解压 ipa 包后得到 xxx.app
	appPath, err := ExtraZip(path)
	if err != nil {
		return nil, err
	}

	//删除解压缓存目录
	defer func() {
		//从xxx.app回退两级获取到xxxx.ipa目录进行删除
		ipaPath := filepath.Dir(filepath.Dir(appPath))
		os.RemoveAll(ipaPath)
	}()
	codesignPath := ExecCodesign(appPath)
	if codesignPath == "" {
		return nil, fmt.Errorf("解析失败")
	}
	feature := &model.Feature{
		Platform: "iOS",
	}
	sha1, modulus := ExecOpenssl(codesignPath)
	feature.PublicKey = modulus
	feature.MD5 = sha1
	label, packageName, icon := ParsePlist(filepath.Join(appPath, "Info.plist"))
	feature.Name = label
	feature.Id = packageName
	feature.Icon = icon

	features := app.Features
	if features == nil {
		features = make([]*model.Feature, 0)
	}
	app.Features = append(features, feature)

	return feature, nil
}

func ParsePlist(path string) (label, packageName, icon string) {
	fp, err := os.Open(path)
	if err != nil {
		return
	}
	defer fp.Close()
	bs, _ := io.ReadAll(fp)
	var result map[string]interface{}
	_, err = plist.Unmarshal(bs, &result)
	if err != nil {
		return
	}
	if bundleId, okB := result["CFBundleIdentifier"]; okB {
		packageName = fmt.Sprintf("%v", bundleId)
	}
	//CFBundleName
	if name, okN := result["CFBundleDisplayName"]; okN {
		label = fmt.Sprintf("%v", name)
	} else if name, okN := result["CFBundleName"]; okN {
		label = fmt.Sprintf("%v", name)
	}

	//CFBundleIcons["CFBundlePrimaryIcon"]["CFBundleIconFiles"]
	if icons, okI := result["CFBundleIcons"]; okI {
		if icons1, okI := icons.(map[string]interface{}); okI {
			if icons2, okI := icons1["CFBundlePrimaryIcon"]; okI {
				if icons2, okI := icons2.(map[string]interface{}); okI {
					if icons3, okI := icons2["CFBundleIconFiles"]; okI {
						if icons4, okI := icons3.([]interface{}); okI {
							if len(icons4) > 0 {
								iconName := fmt.Sprintf("%v", icons4[0])
								parent := filepath.Dir(path)
								var iconPath string
								filepath.Walk(parent, func(path string, info fs.FileInfo, err error) error {
									if strings.Contains(info.Name(), iconName) && strings.HasSuffix(info.Name(), ".png") {
										iconPath = path
									}
									return nil
								})

								fp, err := os.Open(iconPath)
								if err != nil {
									return
								}
								defer fp.Close()
								bs, err := io.ReadAll(fp)
								if err != nil {
									return
								}

								cgbi, err := ipaPng.Decode(bytes.NewReader(bs))
								if err != nil {
									panic(err)
								}

								emptyBuffer := bytes.NewBuffer(nil)                              //开辟一个新的 buffer
								png.Encode(emptyBuffer, cgbi.Img)                                //image 写入 buff
								icon = base64.RawStdEncoding.EncodeToString(emptyBuffer.Bytes()) //buff 转 base64

							}
						}
					}
				}
			}
		}
	}

	return
}

// openssl x509 -fingerprint -sha1 -modulus -text -noout -in codesign0
func ExecOpenssl(path string) (sha1 string, modulus string) {
	parent := filepath.Dir(path)
	cmd := exec.Command("openssl", "x509", "-fingerprint", "-sha1", "-modulus", "-text", "-noout", "-in", path)
	cmd.Dir = parent //设置工作目录，注意不要Path混淆
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	//SplitN的N是截取数组长度
	result := strings.SplitN(string(output), "\n", 3)
	//当长度不为3时，可能是因为结果中没有3个换行符
	if len(result) != 3 {
		return
	}
	sha1 = strings.Split(strings.ReplaceAll(result[0], ":", ""), "=")[1]
	modulus = strings.Split(result[1], "=")[1]
	return
}

func ExecCodesign(path string) string {
	parent := filepath.Dir(path)
	cmd := exec.Command("codesign", "-d", "--extract-certificates", path)
	cmd.Dir = parent //设置工作目录，注意不要Path混淆
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	fmt.Println(string(output))
	return filepath.Join(parent, "codesign0")
}

func ExtraZip(path string) (string, error) {
	ipaInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("文件打开错误")
	}
	ipaName := ipaInfo.Name()
	zipReader, err := zip.OpenReader(path)

	if err != nil {
		return "", fmt.Errorf("文件打开错误")
	}
	defer zipReader.Close()

	const bundleId = "icu.bughub.app.icp_tool"

	cacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", fmt.Errorf("获取缓存目录错误")
	}

	cacheDirPath := filepath.Join(cacheDir, bundleId, ipaName)

	var appPath string

	if Exits(cacheDirPath) {
		os.RemoveAll(cacheDirPath)
	}

	for _, file := range zipReader.File {
		fileInfo := file.FileInfo()

		//路径中不包含payload的不解压，直接continue
		if !strings.Contains(file.Name, "Payload") {
			continue
		}

		filePath := filepath.Join(cacheDirPath, file.Name)

		if fileInfo.IsDir() {
			MkDir(filePath)
			if strings.HasSuffix(fileInfo.Name(), ".app") {
				appPath = filePath
			}
			continue
		}

		//使用匿名函数，解决无法defer
		err := func() error {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			fw, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer fw.Close()
			_, err = io.Copy(fw, rc)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return "", err
		}
	}

	return appPath, nil
}

func Exits(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 如果目录不存在则创建目录
// 成功返回nil，不成功返回目录
func MkDir(path string) error {
	if !Exits(path) {
		err := os.MkdirAll(path, 0755)
		return err
	}
	return nil
}
