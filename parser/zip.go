package parser

import (
	"archive/zip"
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (app *App) SaveToZip(savePath string) error {

	//获取文件名称 /Users/wei/Downloads/备案材料
	var paths = strings.Split(savePath, "/")
	var fileName = paths[len(paths)-1]
	if strings.Contains(fileName, ".") {
		fileName = strings.Split(fileName, ".")[0]
	}

	const bundleId = "icu.bughub.app.icp_tool"

	cacheDir, err := os.UserCacheDir()

	if err != nil {
		return fmt.Errorf("获取缓存目录错误")
	}

	//需要打包的目录
	cacheDirPath := filepath.Join(cacheDir, bundleId, fileName)

	//将内容写入文件
	for _, feature := range app.Features {

		//以名字和平台命名目录
		dirName := fmt.Sprintf("%s%s", feature.Name, feature.Platform)

		dirPath := filepath.Join(cacheDirPath, dirName)

		//打包好之后删除缓存目录
		defer func() {
			os.RemoveAll(dirPath)
		}()

		MkDir(dirPath)

		iconData, err := base64.RawStdEncoding.DecodeString(feature.Icon)
		if err != nil {
			return fmt.Errorf("读取图标失败")
		}
		iconFilePath := filepath.Join(dirPath, "图标.png")

		iconFile, err := os.OpenFile(iconFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("写入图标失败")
		}
		defer iconFile.Close()
		iconFile.Write(iconData)

		//文本文件
		plainFilePath := filepath.Join(dirPath, "材料.txt")
		plainFile, err := os.OpenFile(plainFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("写入图标失败")
		}
		defer plainFile.Close()
		//写入文件时，使用带缓存的 *Writer
		write := bufio.NewWriter(plainFile)
		write.WriteString(fmt.Sprintf("名称：%s\n", feature.Name))
		write.WriteString(fmt.Sprintf("平台：%s\n", feature.Platform))
		write.WriteString(fmt.Sprintf("Bundle Id(Package、包名)：%s\n", feature.Id))
		write.WriteString(fmt.Sprintf("证书MD5指纹(签名MD5值、SHA-1)：%s\n", feature.MD5))
		write.WriteString(fmt.Sprintf("Modulus(公钥)：%s\n", feature.PublicKey))
		//Flush将缓存的文件真正写入到文件中
		write.Flush()
	}

	fw, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer fw.Close()
	// 通过 fw 来创建 zip.Write
	zw := zip.NewWriter(fw)
	defer func() {
		// 检测一下是否成功关闭
		if err := zw.Close(); err != nil {
			return
		}
	}()

	return filepath.Walk(cacheDirPath, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(strings.TrimPrefix(path, filepath.Dir(cacheDirPath)), string(filepath.Separator))

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fr.Close()

		// 将打开的文件 Copy 到 w
		_, err = io.Copy(w, fr)
		if err != nil {
			return err
		}

		return nil

	})

}
