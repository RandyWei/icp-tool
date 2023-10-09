package main

import (
	"embed"
	"fmt"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"icu.bughub.app/ipc-tool/model"
	"icu.bughub.app/ipc-tool/parser"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := parser.NewApp()

	appMenu := menu.NewMenu()

	//只针对MacOS
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
	}

	fileSubMenu := appMenu.AddSubmenu("文件")
	fileSubMenu.AddText("打开文件", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
		filePath, err := wailsRuntime.OpenFileDialog(app.Ctx, wailsRuntime.OpenDialogOptions{
			Title: "打开文件",
			Filters: []wailsRuntime.FileFilter{
				{
					Pattern: "*.ipa;*.apk",
				},
			},
		})
		if err != nil {
			fmt.Printf("err:%T\n", err)
			return
		}

		if strings.TrimSpace(filePath) == "" {
			return
		}

		event := model.Event{
			Ctx:  app.Ctx,
			Name: model.Event_PRRSER,
			Data: model.EventData{
				Status: model.Event_PARSER_LOADING,
			},
		}
		//通知前端文件加载中
		event.Send()

		var feature *model.Feature

		if strings.HasSuffix(filePath, ".apk") {
			feature, err = app.ParseApk(filePath)

		} else {
			feature, err = app.ParseIpa(filePath)
		}

		//得到结果后也需要通知前端
		if err != nil {
			event := model.Event{
				Ctx:  app.Ctx,
				Name: model.Event_PRRSER,
				Data: model.EventData{
					Status: model.Event_PARSER_ERROR,
					Data:   "解析出错",
				},
			}
			event.Send()
			return
		}
		event = model.Event{
			Ctx:  app.Ctx,
			Name: model.Event_PRRSER,
			Data: model.EventData{
				Status: model.Event_PARSER_RESULT,
				Data:   feature,
			},
		}
		event.Send()
	})

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "icp-tool",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             appMenu,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
