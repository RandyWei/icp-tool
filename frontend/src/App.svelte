<script lang="ts">
  import { ParseApk, SaveToZip } from "../wailsjs/go/parser/App.js";
  import type { model } from "wailsjs/go/models";
  import { onMount } from "svelte";
  import * as wailsRuntime from "../wailsjs/runtime/runtime";
  import { toast } from "@zerodevx/svelte-toast";
  import { SvelteToast } from "@zerodevx/svelte-toast";

  enum EventName {
    Parser = "parser",
    Save = "save",
  }

  enum Status {
    Default = "default",
    Loading = "loading",
    Result = "result",
    Error = "error",
  }

  let errorText = "";
  let supportTips = "";
  let currentStatus: Status = Status.Default;
  let features = new Array<model.Feature>();

  function greet(): void {
    currentStatus = Status.Loading;
    ParseApk("/Users/wei/Downloads/com.chinahrt.app.gpjw_1.0.22.apk")
      .then((result) => {
        console.log(result);
        features.push(result);
        currentStatus = Status.Result;
      })
      .catch((e) => {
        console.error(e);
        currentStatus = Status.Error;
      });
  }

  /**
   * 允许拖拽
   * @param event
   */
  var allowDrop = function (event: DragEvent) {
    event.preventDefault();
  };

  /**
   * 拖拽完成
   * @param event
   */
  function drop(event: DragEvent) {
    event.preventDefault();
    let files = event.dataTransfer.files;
    if (files.length > 0) {
      const file = files[0];
      const reader = new FileReader();

      reader.onload = function (e: ProgressEvent<FileReader>) {
        const fileContent = e.target.result as ArrayBuffer;
        console.log(fileContent);
      };
      reader.readAsArrayBuffer(file);
    }
  }

  function initEventListener() {
    //监听解析事件
    wailsRuntime.EventsOn(EventName.Parser, (data: model.EventData) => {
      switch (data.status) {
        case Status.Result:
          currentStatus = Status.Result;
          features.push(data.data);
          break;
        case Status.Loading:
          currentStatus = Status.Loading;
          break;
        case Status.Error:
          currentStatus = Status.Error;
          // errorText = data.data;
          break;
        default:
          currentStatus = Status.Default;
          break;
      }
    });

    wailsRuntime.EventsOn(EventName.Save, (data: model.EventData) => {
      switch (data.status) {
        case "success":
          console.log(data);
          toast.push("保存成功", {
            theme: {
              "--toastBarHeight": 0,
              "--toastColor": "mintcream",
              "--toastBackground": "rgba(0,255,0,0.9)",
              "--toastBarBackground": "red",
            },
          });
          break;
        default:
          toast.push("保存失败", {
            theme: {
              "--toastBarHeight": 0,
              "--toastColor": "mintcream",
              "--toastBackground": "rgba(255,0,0,0.9)",
              "--toastBarBackground": "red",
            },
          });
          break;
      }
    });
  }

  onMount(function () {
    //注册拖拽事件
    // let dropArea = document.querySelector("#container");
    // dropArea.addEventListener("drop", drop);
    // dropArea.addEventListener("dragover", allowDrop);

    initEventListener();

    wailsRuntime.Environment().then((info) => {
      if (info.platform == "darwin") {
        supportTips = "当前系统支持解析ipa和apk包";
      } else if (info.platform == "windows") {
        supportTips = "当前系统支持解析apk包";
      }
    });
  });
</script>

<main>
  <SvelteToast />
  <div
    style="color: red;background-color:bisque;width:100%;position:sticky;top:0;margin-bottom: 1.5rem;"
  >
    {supportTips}
  </div>
  <div id="container">
    {#if currentStatus == Status.Default}
      <div id="tip">
        通过文件菜单->打开文件选择apk或ipa进行解析，也可以使用Ctrl/Command +
        O快捷键
      </div>
    {:else if currentStatus == Status.Loading}
      <div id="tip">正在解析中</div>
    {:else if currentStatus == Status.Error}
      <div id="tip">解析失败，请重新尝试{errorText}</div>
    {:else}
      {#each features as apkFeature}
        <div id="result">
          <div class="line">
            <img
              src="data:image/png;base64,{apkFeature.icon}"
              alt=""
              style="max-width: 5rem;max-height:5rem"
            />
          </div>
          <div class="line">APP名称：{apkFeature.name}</div>
          <div class="line">平台：{apkFeature.platform}</div>
          <div class="line">Bundle Id：{apkFeature.id}</div>
          <div class="line">
            证书MD5指纹(签名MD5值、sha-1)：{apkFeature.md5}
          </div>
          <div class="line">Modulus(公钥)：{apkFeature.publicKey}</div>
          <hr />
        </div>
      {/each}
    {/if}
  </div>
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100%;
    min-height: 100%;
  }
  #container {
    width: 80%;
    height: 100%;
    min-height: 100%;
    text-align: left;
    word-break: break-all;
  }
  #tip {
    font-size: xx-large;
    height: 100%;
    max-height: 100%;
    text-align: center;
    align-items: center;
    justify-content: center;
    display: flex;
  }
  .line {
    padding: 0.5rem 0;
  }
</style>
