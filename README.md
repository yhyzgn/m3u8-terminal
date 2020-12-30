# `m3u8`下载器
![release](https://img.shields.io/github/v/release/yhyzgn/m3u8?color=brightgreen)



## 推荐

> 由于采用`ffmpeg`合并和转换媒体，所以必须搭建`ffmpeg`环境

* 环境变量

  > 优先使用环境变量中的`ffmpeg`

* 同级目录

  > 其次选择同级目录下的`ffmpeg`

* 远程下载

  > 如果环境变量和同级目录下都不存在`ffmpeg`，程序启动时将自动根据配置文件`settings.json`所配置的下载源下载到程序同级目录



## 配置文件

> 当前目录下的`settings.json`文件

```json
{
  "//": "媒体保存目录",
  "saveDir": "./Download",
  "//": "临时分片ts文件保存目录前缀，临时目录命名：${tsTempDirPrefix} + ${MD5(url)}",
  "tsTempDirPrefix": "ts_",
  "//": "默认媒体格式",
  "extension": "mp4",
  "//": "各平台 ffmpeg 下载源，直接下载可执行程序",
  "ffmpeg": {
    "windows": "https://gitee.com/yhyzgn/ffmpeg/attach_files/566856/download/ffmpeg.exe",
    "mac": "https://gitee.com/yhyzgn/ffmpeg/attach_files/566856/download/ffmpeg.mac",
    "linux": "https://gitee.com/yhyzgn/ffmpeg/attach_files/566856/download/ffmpeg.linux"
  }
}
```



## 运行下载

> 命令行模式

```shell
./m3u8 -url "https://src.com/index.m3u8" -name "测试" -ext "mp4"
```

参数说明

| 参数名 |  可空   |                  说明                  |
| :----: | :-----: | :------------------------------------: |
| `url`  | `false` |               `m3u8`链接               |
| `name` | `true`  |     媒体文件名，空时使用`MD5(url)`     |
| `ext`  | `true`  | 媒体类型，空时使用`settings.extension` |

