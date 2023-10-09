package parser

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/avast/apkparser"
	"github.com/avast/apkverifier"
	"icu.bughub.app/ipc-tool/model"
)

func (app *App) Test() *model.EventData {
	return nil
}

func (app *App) ParseApk(path string) (*model.Feature, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("文件打开错误")
	}
	defer f.Close()
	optionalZip, err := apkparser.OpenZipReader(f)
	if err != nil {
		return nil, fmt.Errorf("文件打开错误")
	}

	//解析manifest
	packageName, name, iconPath, err := getManifestInfo(optionalZip)
	if err != nil {
		return nil, fmt.Errorf("解析错误")
	}

	res, err := apkverifier.Verify(path, optionalZip)
	if err != nil {
		return nil, fmt.Errorf("解析错误")
	}
	certInfo, certs := apkverifier.PickBestApkCert(res.SignerCerts)

	feature := new(model.Feature)

	if publicKey, ok := certs.PublicKey.(*rsa.PublicKey); ok {
		feature.PublicKey = strings.ToUpper(publicKey.N.String())
	}
	feature.MD5 = strings.ToUpper(certInfo.Md5)
	feature.Icon = getIconData(optionalZip, iconPath)
	feature.Id = packageName
	feature.Name = name
	return feature, nil
}

func getIconData(zip *apkparser.ZipReader, iconPath string) string {
	icon := zip.File[iconPath]
	if icon == nil {
		return ""
	}

	if err := icon.Open(); err != nil {
		return ""
	}

	defer icon.Close()

	data := make([]byte, 0)
	buf := make([]byte, 4)
	for {
		_, err := icon.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		data = append(data, buf...)
	}

	return base64.RawStdEncoding.EncodeToString(data)

}

func getManifestInfo(zip *apkparser.ZipReader) (packageName string, name string, iconPath string, err error) {

	resources := zip.File["resources.arsc"]

	if resources == nil {
		err = fmt.Errorf("打开resources文件失败", err.Error())
		return
	}

	if err = resources.Open(); err != nil {
		err = fmt.Errorf("打开resources文件失败", err.Error())
		return
	}

	defer resources.Close()

	resourceTable, err := apkparser.ParseResourceTable(resources)

	if err != nil {
		err = fmt.Errorf("打开resources文件失败", err.Error())
		return
	}

	manifest := zip.File["AndroidManifest.xml"]
	if manifest == nil {
		err = fmt.Errorf("打开manifest文件失败", err.Error())
		return
	}

	if err = manifest.Open(); err != nil {
		err = fmt.Errorf("打开manifest文件失败", err.Error())
		return
	}

	defer manifest.Close()

	for manifest.Next() {
		encoder := manifestEncoder{}
		if err = apkparser.ParseXml(manifest, &encoder, resourceTable); err != nil {
			err = fmt.Errorf("打开manifest文件失败", err.Error())
			return
		}
		return encoder.packageName, encoder.name, encoder.iconPath, nil
	}
	return
}

type manifestEncoder struct {
	packageName string
	name        string
	iconPath    string
}

func (e *manifestEncoder) EncodeToken(t xml.Token) error {
	st, ok := t.(xml.StartElement)
	if !ok {
		return nil
	}

	switch st.Name.Local {
	case "manifest":
		val := e.getAttrStringValue(&st, "package")
		e.packageName = val
	case "application":
		name := e.getAttrStringValue(&st, "label")
		icon := e.getAttrStringValue(&st, "icon")
		e.name = name
		e.iconPath = icon
	}
	return nil
}

func (e *manifestEncoder) Flush() error {
	return nil
}

func (e *manifestEncoder) getAttrIntValue(st *xml.StartElement, name string) (int32, error) {
	for _, attr := range st.Attr {
		if attr.Name.Local == name {
			val, err := strconv.ParseInt(attr.Value, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("failed to decode %s '%s': %s", name, attr.Value, err.Error())
			}
			return int32(val), nil
		}
	}
	return 0, io.EOF
}

func (e *manifestEncoder) getAttrStringValue(st *xml.StartElement, name string) string {
	for _, attr := range st.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}
