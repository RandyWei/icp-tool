# 获取 APP 特征信息

可以快速获取 APP 的特征信息以供备案使用

在 Mac OS 上可以支持获取 apk 和 ipa 的信息，在 Windows 上只支持获取 apk 信息

参考链接：

[https://help.aliyun.com/zh/icp-filing/user-guide/fill-in-app-feature-information?spm=a2cmq.17629970.icp_beian.16.f0d079feGzvc9w](https://help.aliyun.com/zh/icp-filing/user-guide/fill-in-app-feature-information?spm=a2cmq.17629970.icp_beian.16.f0d079feGzvc9w)

[https://cloud.tencent.com/document/product/243/97789](https://cloud.tencent.com/document/product/243/97789)

## 使用方法

1. 使用打开文件(Ctrl/Command + O)菜单选择即将发布到 app store 的 ipa 或 release 版本的 apk
2. 等待解析好之后就会显示特征信息
3. 选择保存为 zip(Ctrl/Command + S)菜单进行保存

## 本地运行项目

> > 前提条件：安装了 wails

```
wails dev
```

## 原理

### iOS

- 解压 ipa 包后得到 xxx.app

- 执行以下命令可将 xxx.app 得到签名文件

```
codesign -d --extract-certificates xxxx.app
```

- 执行以下命令获取得 sha-1 和 modulus 信息

```
openssl x509 -fingerprint -sha1 -modulus -text -noout -in codesign0
```

- 解析 xxx.app 中的 Info.plist 文件得到 bundle id 和 名称以及图标路径

- 使用以下合集还原图片以保证在其他终端正常显示

```
xcrun -sdk iphoneos pngcrush \ -q -revert-iphone-optimizations -d AppIcon60x60@2x.png
```

### Android

通过开源库 github.com/avast/apkparser 解析
