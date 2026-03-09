# Qiniu Kling Provider

七牛云 Kling API 后端实现，支持所有 Kling 图像和视频生成模型。

## 概述

本包实现了 `kling.ImageGenExecutor` 和 `kling.VideoGenExecutor` 接口，通过七牛云 API 提供 Kling 模型的图像和视频生成能力。

## 支持的模型

### 图像模型

| 模型 | 能力 | 说明 |
|------|------|------|
| `kling-v1` | 文生图、单图生图 | 基础版本 |
| `kling-v1-5` | 文生图、单图生图 | 支持角色特征/人物长相参考 |
| `kling-v2` | 文生图、单图生图、多图生图 | 支持 subject_image_list |
| `kling-v2-new` | 单图生图 | 仅支持图生图 |
| `kling-v2-1` | 文生图、单图生图、多图生图 | 最新版本 |
| `kling-image-o1` | 文生图、多模态 | OmniImage，支持 `<<<image_1>>>` 引用 |

### 视频模型

| 模型 | 能力 | 说明 |
|------|------|------|
| `kling-v2-1` | 图生视频、首尾帧生视频 | 必须提供 input_reference |
| `kling-v2-5-turbo` | 文生视频、图生视频、首尾帧生视频 | 全功能 |
| `kling-v2-6` ~ `kling-v2-9` | 文生视频、图生视频、有声视频、动作控制 | 支持 sound 参数 |
| `kling-video-o1` | 文生视频、图生视频、视频生视频 | 使用 image_list/video_list |
| `kling-v3` | 文生视频、图生视频、有声视频 | 最新版本 |
| `kling-v3-omni` | 全功能 | 支持多镜头、multi_prompt |

## 快速开始

### 方式一：使用 Register 函数

```go
import (
    "github.com/goplus/xai/spec/kling/provider/qiniu"
)

// 注册 qiniu 后端到全局 xai
qiniu.Register("your-api-token")

// 然后通过 xai.New 使用
svc, _ := xai.New(ctx, "kling://")
```

### 方式二：手动创建 Service

```go
import (
    "github.com/goplus/xai/spec/kling"
    "github.com/goplus/xai/spec/kling/provider/qiniu"
)

// 创建 service
svc := qiniu.NewService("your-api-token")

// 或者使用自定义选项
svc := qiniu.NewService("your-api-token",
    qiniu.WithBaseURL(qiniu.OverseasBaseURL),  // 海外端点
    qiniu.WithUserID("user-123"),               // 用户追踪
)
```

### 方式三：分别创建 Executor

```go
client := qiniu.NewClient("your-api-token")
imgExec := qiniu.NewImageExecutor(client)
vidExec := qiniu.NewVideoExecutor(client)

svc := kling.NewService(imgExec, vidExec)
```

## 使用示例

### 图像生成

```go
ctx := context.Background()
svc := qiniu.NewService(token)

// 获取图像生成操作
op, err := svc.Operation(kling.ModelKlingV2, xai.GenImage)
if err != nil {
    log.Fatal(err)
}

// 设置参数
op.Params().
    Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳").
    Set(kling.ParamAspectRatio, kling.Aspect16x9).
    Set(kling.ParamN, 2)

// 调用并等待结果
results, err := xai.Call(ctx, svc, op, nil, nil)
if err != nil {
    log.Fatal(err)
}

// 获取生成的图像
for i := 0; i < results.Len(); i++ {
    img := results.At(i).(*xai.OutputImage)
    fmt.Printf("Image %d: %s\n", i, img.Image.StgUri())
}
```

### 多图生图 (kling-v2/v2-1)

```go
op, _ := svc.Operation(kling.ModelKlingV2, xai.GenImage)
op.Params().
    Set(kling.ParamPrompt, "综合两个图像画一个漫画图").
    Set(kling.ParamSubjectImageList, []map[string]string{
        {"subject_image": "https://example.com/img1.jpg"},
        {"subject_image": "https://example.com/img2.jpg"},
    }).
    Set(kling.ParamAspectRatio, kling.Aspect16x9)
```

### O1 图像生成

```go
op, _ := svc.Operation(kling.ModelKlingImageO1, xai.GenImage)
op.Params().
    Set(kling.ParamPrompt, "参考 <<<image_1>>> 的风格，增加一群人").
    Set(kling.ParamReferenceImages, []string{"https://example.com/ref.jpg"}).
    Set(kling.ParamResolution, kling.Resolution2K).
    Set(kling.ParamN, 2)
```

### 视频生成

```go
op, _ := svc.Operation(kling.ModelKlingV3, xai.GenVideo)
op.Params().
    Set(kling.ParamPrompt, "一只可爱的小猫在阳光下玩耍").
    Set(kling.ParamMode, kling.ModePro).
    Set(kling.ParamSeconds, "5").
    Set(kling.ParamSound, kling.SoundOn)

results, _ := xai.Call(ctx, svc, op, nil, func(resp xai.OperationResponse) {
    fmt.Printf("Task %s processing...\n", resp.TaskID())
})
```

### 图生视频

```go
op, _ := svc.Operation(kling.ModelKlingV21Video, xai.GenVideo)
op.Params().
    Set(kling.ParamPrompt, "人在奔跑").
    Set(kling.ParamInputReference, "https://example.com/first-frame.jpg").
    Set(kling.ParamMode, kling.ModePro).
    Set(kling.ParamSize, kling.Size1920x1080)
```

### 首尾帧生视频

```go
op, _ := svc.Operation(kling.ModelKlingV25Turbo, xai.GenVideo)
op.Params().
    Set(kling.ParamPrompt, "人跑到了天涯海角").
    Set(kling.ParamInputReference, "https://example.com/first.jpg").
    Set(kling.ParamImageTail, "https://example.com/last.jpg").
    Set(kling.ParamMode, kling.ModePro).
    Set(kling.ParamSeconds, kling.Seconds10)
```

## API 端点

| 端点 | 说明 |
|------|------|
| `https://api.qnaigc.com` | 国内端点 (默认) |
| `https://openai.sufy.com` | 海外端点 |

## 配置选项

```go
// 使用海外端点
qiniu.WithBaseURL(qiniu.OverseasBaseURL)

// 设置用户 ID (用于追踪)
qiniu.WithUserID("user-123")

// 自定义 HTTP 客户端
qiniu.WithHTTPClient(&http.Client{Timeout: 120 * time.Second})
```

## 异步任务处理

所有生成任务都是异步的，返回 `xai.OperationResponse`：

```go
// 提交任务
resp, err := op.Call(ctx, svc, nil)

// 检查是否完成
if resp.Done() {
    results := resp.Results()
    // 处理结果
} else {
    // 任务仍在处理中，获取 task ID
    taskID := resp.TaskID()
    
    // 轮询等待
    resp.Sleep()
    resp, err = resp.Retry(ctx, svc)
}

// 或使用 xai.Wait 自动轮询
results, err := xai.Wait(ctx, svc, resp, func(r xai.OperationResponse) {
    fmt.Printf("Processing: %s\n", r.TaskID())
})
```

## 错误处理

```go
results, err := xai.Call(ctx, svc, op, nil, nil)
if err != nil {
    var apiErr *qiniu.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s\n", apiErr.Error())
    }
    if errors.Is(err, qiniu.ErrTaskFailed) {
        fmt.Println("Task failed")
    }
}
```

## 文件结构

```
spec/kling/provider/qiniu/
├── client.go          # HTTP 客户端
├── executor.go        # ImageExecutor 和 VideoExecutor
├── image.go           # 图像 API 请求构建
├── video.go           # 视频 API 请求构建
├── response.go        # API 响应类型
├── qiniu_test.go      # 单元测试
├── kling_image.md     # 图像 API 文档
├── kling_video.md     # 视频 API 文档
└── README.md          # 本文件
```

## 参考文档

- [七牛云 Qnagic API](https://apidocs.qnaigc.com)
- [kling_image.md](./kling_image.md) - 图像 API 详细文档
- [kling_video.md](./kling_video.md) - 视频 API 详细文档
