export interface ApiParam {
  name: string
  type: string
  required: boolean
  description: string
}

export interface ApiEndpoint {
  id: string
  method: string
  path: string
  title: string
  description: string
  params?: ApiParam[]
  queryParams?: ApiParam[]
  bodyParams?: ApiParam[]
  responseExample?: string
  /** 是否可在线测试，默认 true；说明类设为 false */
  testable?: boolean
  /** 发送前的警告提示 */
  dangerWarning?: string
}

export interface ApiSection {
  id: string
  title: string
  description?: string
  endpoints: ApiEndpoint[]
}

export const apiSections: ApiSection[] = [
  {
    id: 'videos',
    title: '视频任务',
    description: '视频生成任务的创建、查询和下载。认证方式：Authorization: Bearer <API_KEY>。',
    endpoints: [
      {
        id: 'create-video',
        method: 'POST',
        path: '/v1/videos',
        title: '创建视频任务',
        dangerWarning: '此操作会创建真实任务并消耗账号配额，确认发送？',
        description: '提交视频生成任务。通过 model 参数指定画质、方向和时长，通过 input_reference 传入参考图片 URL 实现图生视频。',
        bodyParams: [
          { name: 'model', type: 'string', required: true, description: '模型名称，如 sora-2-landscape-10s' },
          { name: 'prompt', type: 'string', required: true, description: '视频生成提示词' },
          { name: 'input_reference', type: 'string', required: false, description: '参考图片 URL（图生视频，作为首帧引导）' },
        ],
        responseExample: `{
  "id": "task_a1b2c3d4",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "queued",
  "progress": 0,
  "created_at": 1709251234,
  "size": "1280x720"
}`,
      },
      {
        id: 'get-video',
        method: 'GET',
        path: '/v1/videos/:id',
        title: '查询任务状态',
        description: '根据任务 ID 查询当前状态和进度。轮询此接口直到 status 变为 completed 或 failed。',
        params: [
          { name: 'id', type: 'string', required: true, description: '任务 ID（如 task_a1b2c3d4）' },
        ],
        responseExample: `{
  "id": "task_a1b2c3d4",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "completed",
  "progress": 100,
  "created_at": 1709251234,
  "size": "1280x720"
}`,
      },
      {
        id: 'download-video',
        method: 'GET',
        path: '/v1/videos/:id/content',
        title: '下载视频',
        description: '下载已完成任务的视频文件。仅当 status 为 completed 时可用，返回 video/mp4 二进制流。',
        params: [
          { name: 'id', type: 'string', required: true, description: '任务 ID（如 task_a1b2c3d4）' },
        ],
        responseExample: `// Content-Type: video/mp4
// 返回视频二进制流`,
      },
    ],
  },
  {
    id: 'models',
    title: '可用模型',
    description: '以下为所有支持的模型名称和对应参数。创建任务时 model 字段必须使用这些名称。',
    endpoints: [
      {
        id: 'model-list',
        method: 'GET',
        path: '模型速查表',
        title: '模型速查表',
        testable: false,
        description: `支持两种命名格式（等价）：\`sora-2-xxx\` 和 \`sora_video2-xxx\`

**标准画质 (720p)**

| 模型名 | 方向 | 时长 | 分辨率 |
|--------|------|------|--------|
| sora-2-landscape-10s | 横屏 | 10s | 1280x720 |
| sora-2-landscape-15s | 横屏 | 15s | 1280x720 |
| sora-2-landscape-25s | 横屏 | 25s | 1280x720 |
| sora-2-portrait-10s | 竖屏 | 10s | 720x1280 |
| sora-2-portrait-15s | 竖屏 | 15s | 720x1280 |
| sora-2-portrait-25s | 竖屏 | 25s | 720x1280 |

**高清画质 (1080p) — Pro**

| 模型名 | 方向 | 时长 | 分辨率 |
|--------|------|------|--------|
| sora-2-pro-landscape-hd-10s | 横屏 | 10s | 1920x1080 |
| sora-2-pro-landscape-hd-15s | 横屏 | 15s | 1920x1080 |
| sora-2-pro-landscape-hd-25s | 横屏 | 25s | 1920x1080 |
| sora-2-pro-portrait-hd-10s | 竖屏 | 10s | 1080x1920 |
| sora-2-pro-portrait-hd-15s | 竖屏 | 15s | 1080x1920 |
| sora-2-pro-portrait-hd-25s | 竖屏 | 25s | 1080x1920 |`,
      },
    ],
  },
  {
    id: 'errors',
    title: '状态码与错误',
    description: 'API 使用标准 HTTP 状态码，错误响应格式统一。',
    endpoints: [
      {
        id: 'error-codes',
        method: 'GET',
        path: '参考说明',
        title: '状态码说明',
        testable: false,
        description: `**HTTP 状态码**

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误（如模型名无效） |
| 401 | 认证失败（API Key 无效或已禁用） |
| 404 | 任务不存在 |
| 503 | 无可用账号 |

**任务状态**

| 状态 | 说明 |
|------|------|
| queued | 已提交，等待处理 |
| in_progress | 生成中 |
| completed | 已完成，可下载 |
| failed | 失败 |

**错误响应格式**

\`{ "error": { "message": "错误描述" } }\``,
      },
    ],
  },
]
