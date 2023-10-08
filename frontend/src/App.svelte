<script lang="ts">
  import logo from "./assets/images/logo-universal.png";
  import { Parse } from "../wailsjs/go/parser/App.js";
  import type { model } from "wailsjs/go/models";

  let resultText: string = "Please enter your name below 👇";
  let name: string;

  enum Status {
    Default,
    Loading,
    Result,
    Error,
  }

  let currentStatus: Status = Status.Default;
  let apkFeature: model.Feature;

  function greet(): void {
    currentStatus = Status.Loading;
    Parse("/Users/wei/Downloads/com.chinahrt.app.gpjw_1.0.22.apk")
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

  function saveToZip() {}
</script>

<main>
  <div class="input-box" id="input">
    <button class="btn" on:click={greet}>Greet</button>
  </div>
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
