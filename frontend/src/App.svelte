<script lang="ts">
  import { ParseApk } from "../wailsjs/go/parser/App.js";
  import type { model } from "wailsjs/go/models";
  import { onMount } from "svelte";
  import * as wailsRuntime from "../wailsjs/runtime/runtime";

  enum EventName {
    Parser = "parser",
  }

  enum Status {
    Default = "default",
    Loading = "loading",
    Result = "result",
    Error = "error",
  }

  let currentStatus: Status = Status.Default;
  let apkFeature: model.Feature;

  function greet(): void {
    currentStatus = Status.Loading;
    ParseApk("/Users/wei/Downloads/com.chinahrt.app.gpjw_1.0.22.apk")
      .then((result) => {
        console.log(result);
        apkFeature = result;
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

  function saveToZip() {}

  function initEventListener() {
    //监听解析事件
    wailsRuntime.EventsOn(EventName.Parser, (data: model.EventData) => {
      switch (data.status) {
        case Status.Result:
          currentStatus = Status.Result;
          apkFeature = data.data;
          break;
        case Status.Loading:
          currentStatus = Status.Loading;
          break;
        case Status.Error:
          currentStatus = Status.Error;
          break;
        default:
          currentStatus = Status.Default;
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
  });
</script>

<main>
  <div id="container">
    {#if currentStatus == Status.Default}
      <div id="tip">请将ipa包拖进来</div>
    {:else if currentStatus == Status.Loading}
      <div id="tip">正在解析中</div>
    {:else if currentStatus == Status.Error}
      <div id="tip">解析失败，请重新尝试</div>
    {:else}
      <div id="result">
        <div class="line">
          <img src="data:image/png;base64,{apkFeature.icon}" alt="" />
        </div>
        <div class="line">APP名称：{apkFeature.name}</div>
        <div class="line">Bundle Id：{apkFeature.id}</div>
        <div class="line">证书MD5指纹(签名MD5值、sha-1)：{apkFeature.md5}</div>
        <div class="line">Modulus(公钥)：{apkFeature.publicKey}</div>
      </div>
      <div id="save-container">
        <button on:click={saveToZip}>保存为zip</button>
      </div>
    {/if}
  </div>
</main>

<style>
  main {
    display: flex;
    justify-content: center;
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
    height: 100vh;
    text-align: center;
    align-items: center;
    justify-content: center;
    display: flex;
  }
  .line {
    padding: 0.5rem 0;
  }
  #save-container {
    width: 100%;
    text-align: center;
    padding-bottom: 0.5rem;
  }
  button {
    width: 50%;
    height: 2rem;
  }
</style>
