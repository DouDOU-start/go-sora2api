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
  /** 所属分组标识（用于 UI 分组展示） */
  group?: string
  endpoints: ApiEndpoint[]
}

/** API 分组定义 */
export interface ApiGroup {
  id: string
  title: string
}

/** 分组列表（按展示顺序） */
export const apiGroupDefs: ApiGroup[] = [
  { id: 'getting-started', title: '入门' },
  { id: 'video', title: '视频' },
  { id: 'image', title: '图片' },
  { id: 'manage', title: '管理' },
  { id: 'tools', title: '工具' },
  { id: 'reference', title: '参考' },
]

export const apiSections: ApiSection[] = [
  // ── 认证 ──
  {
    id: 'auth',
    title: '认证',
    group: 'getting-started',
    description: 'API 请求需通过 API Key 进行身份认证。',
    endpoints: [
      {
        id: 'auth-info',
        method: 'GET',
        path: '参考说明',
        title: '认证方式',
        testable: false,
        description: `所有 \`/v1/\` 接口需要在请求头中携带 API Key：

\`Authorization: Bearer <your-api-key>\`

API Key 可在管理后台的「密钥」页面创建和管理。

**错误响应**

认证失败时返回 401：\`{ "error": { "message": "无效的 API Key" } }\``,
      },
    ],
  },
  // ── 视频任务 ──
  {
    id: 'videos',
    title: '视频生成',
    group: 'video',
    description: '文生视频和图生视频。通过 style 参数或 prompt 中的 {style} 标记可指定风格。',
    endpoints: [
      {
        id: 'create-video',
        method: 'POST',
        path: '/v1/videos',
        title: '创建视频任务',
        dangerWarning: '此操作会创建真实任务并消耗账号配额，确认发送？',
        description: `提交视频生成任务。支持文生视频（仅 model + prompt）和图生视频（额外传 input_reference）。`,
        bodyParams: [
          { name: 'model', type: 'string', required: true, description: '模型名称，如 sora-2-landscape-10s' },
          { name: 'prompt', type: 'string', required: true, description: '视频生成提示词' },
          { name: 'input_reference', type: 'string', required: false, description: '参考图片 URL 或 base64 data URI（图生视频，支持 PNG/JPEG/WebP）' },
          { name: 'style', type: 'string', required: false, description: '视频风格（如 anime, retro, comic 等，见模型速查表）' },
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
        title: '查询视频任务状态',
        description: '根据任务 ID 查询当前状态和进度。轮询此接口直到 status 变为 completed 或 failed。',
        params: [
          { name: 'id', type: 'string', required: true, description: '任务 ID（如 task_a1b2c3d4）' },
        ],
        responseExample: `// 成功
{
  "id": "task_a1b2c3d4",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "completed",
  "progress": 100,
  "created_at": 1709251234,
  "size": "1280x720"
}

// 失败
{
  "id": "task_a1b2c3d4",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "failed",
  "progress": 30,
  "created_at": 1709251234,
  "size": "1280x720",
  "error": { "message": "生成失败: content violation" }
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
  // ── Remix 视频 ──
  {
    id: 'remix',
    title: 'Remix 视频',
    group: 'video',
    description: '基于已有 Sora 视频进行二次创作。需要提供原视频的分享链接或 ID。',
    endpoints: [
      {
        id: 'create-remix',
        method: 'POST',
        path: '/v1/videos/remix',
        title: '创建 Remix 任务',
        dangerWarning: '此操作会创建真实任务并消耗账号配额，确认发送？',
        description: '基于已有 Sora 视频重新创作。remix_target 支持 Sora 分享链接（如 `https://sora.com/g/s_xxx`）或直接传 `s_xxx` 格式 ID。',
        bodyParams: [
          { name: 'model', type: 'string', required: true, description: '模型名称（决定时长和分辨率）' },
          { name: 'prompt', type: 'string', required: true, description: '重新创作的提示词' },
          { name: 'remix_target', type: 'string', required: true, description: 'Sora 分享链接或 s_xxx 格式视频 ID' },
          { name: 'style', type: 'string', required: false, description: '视频风格' },
        ],
        responseExample: `{
  "id": "task_e5f6g7h8",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "queued",
  "progress": 0,
  "created_at": 1709251234,
  "size": "1280x720"
}`,
      },
    ],
  },
  // ── 分镜视频 ──
  {
    id: 'storyboard',
    title: '分镜视频',
    group: 'video',
    description: '通过分镜脚本生成多段式视频。使用 [Ns] 格式定义每个镜头的时长和内容。',
    endpoints: [
      {
        id: 'create-storyboard',
        method: 'POST',
        path: '/v1/videos/storyboard',
        title: '创建分镜任务',
        dangerWarning: '此操作会创建真实任务并消耗账号配额，确认发送？',
        description: `使用分镜脚本生成视频。prompt 使用 \`[Ns]\` 格式定义每个镜头：

\`[5.0s]猫在草地上奔跑 [5.0s]猫跳上围墙眺望远方\`

可在分镜前添加总体描述作为全局指导。`,
        bodyParams: [
          { name: 'model', type: 'string', required: true, description: '模型名称（决定时长和分辨率）' },
          { name: 'prompt', type: 'string', required: true, description: '分镜格式提示词，如 [5.0s]场景1 [5.0s]场景2' },
          { name: 'input_reference', type: 'string', required: false, description: '参考图片 URL 或 base64 data URI' },
          { name: 'style', type: 'string', required: false, description: '视频风格' },
        ],
        responseExample: `{
  "id": "task_i9j0k1l2",
  "object": "video",
  "model": "sora-2-landscape-10s",
  "status": "queued",
  "progress": 0,
  "created_at": 1709251234,
  "size": "1280x720"
}`,
      },
    ],
  },
  // ── 图片任务 ──
  {
    id: 'images',
    title: '图片任务',
    group: 'image',
    description: '图片生成任务的创建、查询和下载。支持文生图和图生图。',
    endpoints: [
      {
        id: 'create-image',
        method: 'POST',
        path: '/v1/images',
        title: '创建图片任务',
        dangerWarning: '此操作会创建真实任务并消耗账号配额，确认发送？',
        description: '提交图片生成任务。支持文生图（仅 prompt）和图生图（额外传 input_reference，支持 URL 或 base64 data URI）。',
        bodyParams: [
          { name: 'prompt', type: 'string', required: true, description: '图片生成提示词' },
          { name: 'width', type: 'integer', required: false, description: '图片宽度（默认 1792）' },
          { name: 'height', type: 'integer', required: false, description: '图片高度（默认 1024）' },
          { name: 'input_reference', type: 'string', required: false, description: '参考图片 URL 或 base64 data URI（图生图）' },
        ],
        responseExample: `{
  "id": "task_a1b2c3d4",
  "object": "image",
  "status": "queued",
  "progress": 0,
  "created_at": 1709251234,
  "width": 1792,
  "height": 1024
}`,
      },
      {
        id: 'get-image',
        method: 'GET',
        path: '/v1/images/:id',
        title: '查询图片任务状态',
        description: '根据任务 ID 查询图片任务状态。完成后返回 image_url。',
        params: [
          { name: 'id', type: 'string', required: true, description: '任务 ID' },
        ],
        responseExample: `{
  "id": "task_a1b2c3d4",
  "object": "image",
  "status": "completed",
  "progress": 100,
  "created_at": 1709251234,
  "width": 1792,
  "height": 1024,
  "image_url": "https://..."
}`,
      },
      {
        id: 'download-image',
        method: 'GET',
        path: '/v1/images/:id/content',
        title: '下载图片',
        description: '下载已完成的图片文件。仅当 status 为 completed 时可用。',
        params: [
          { name: 'id', type: 'string', required: true, description: '任务 ID' },
        ],
        responseExample: `// Content-Type: image/png 或 image/webp
// 返回图片二进制流`,
      },
    ],
  },
  // ── 角色管理 ──
  {
    id: 'characters',
    title: '角色管理',
    group: 'manage',
    description: '创建和管理 Sora 角色。上传包含人物的视频，系统自动完成处理和定稿。',
    endpoints: [
      {
        id: 'create-character',
        method: 'POST',
        path: '/v1/characters',
        title: '创建角色',
        dangerWarning: '此操作会上传视频并创建角色，确认发送？',
        description: '上传包含人物的视频创建角色。后台自动完成：上传 → 处理 → 定稿。轮询查询接口获取最终状态。',
        bodyParams: [
          { name: 'video_url', type: 'string', required: true, description: '角色视频 URL 或 base64 data URI（mp4 格式）' },
          { name: 'username', type: 'string', required: false, description: '角色用户名（不传则使用系统推荐值）' },
          { name: 'display_name', type: 'string', required: false, description: '角色显示名称（不传则使用系统推荐值）' },
        ],
        responseExample: `{
  "id": "char_a1b2c3d4",
  "status": "processing",
  "created_at": 1709251234
}`,
      },
      {
        id: 'get-character',
        method: 'GET',
        path: '/v1/characters/:id',
        title: '查询角色状态',
        description: '查询角色的处理状态。status 为 ready 时表示角色已就绪。',
        params: [
          { name: 'id', type: 'string', required: true, description: '角色 ID（如 char_a1b2c3d4）' },
        ],
        responseExample: `{
  "id": "char_a1b2c3d4",
  "status": "ready",
  "display_name": "John",
  "username": "john_doe",
  "profile_url": "https://...",
  "character_id": "char_xxx",
  "created_at": 1709251234
}`,
      },
      {
        id: 'set-character-public',
        method: 'POST',
        path: '/v1/characters/:id/public',
        title: '设置角色公开',
        description: '将角色设为公开可见。仅当角色 status 为 ready 时可用。',
        params: [
          { name: 'id', type: 'string', required: true, description: '角色 ID' },
        ],
        responseExample: `{
  "message": "角色已设为公开"
}`,
      },
      {
        id: 'delete-character',
        method: 'DELETE',
        path: '/v1/characters/:id',
        title: '删除角色',
        description: '删除角色。如果角色已定稿，同时从 Sora 平台删除。',
        params: [
          { name: 'id', type: 'string', required: true, description: '角色 ID' },
        ],
        responseExample: `// 204 No Content`,
      },
    ],
  },
  // ── 提示词工具 ──
  {
    id: 'prompt',
    title: '提示词工具',
    group: 'tools',
    description: '使用 Sora AI 优化和增强提示词。',
    endpoints: [
      {
        id: 'enhance-prompt',
        method: 'POST',
        path: '/v1/enhance-prompt',
        title: '提示词增强',
        description: '使用 Sora 的 AI 提示词优化能力增强原始提示词，生成更详细、更适合视频生成的描述。',
        bodyParams: [
          { name: 'prompt', type: 'string', required: true, description: '原始提示词' },
          { name: 'expansion_level', type: 'string', required: false, description: '扩展级别：medium 或 long（默认 medium）' },
          { name: 'duration', type: 'integer', required: false, description: '目标视频时长：5/10/15/25 秒（默认 10）' },
        ],
        responseExample: `{
  "original_prompt": "一只猫在花园里",
  "enhanced_prompt": "A fluffy orange tabby cat gracefully walks through a sunlit garden..."
}`,
      },
    ],
  },
  // ── 帖子管理 ──
  {
    id: 'posts',
    title: '帖子管理',
    group: 'manage',
    description: '发布和删除 Sora 视频帖子。发布后可通过 Sora 平台分享。',
    endpoints: [
      {
        id: 'publish-post',
        method: 'POST',
        path: '/v1/posts',
        title: '发布视频帖子',
        dangerWarning: '此操作会将视频发布到 Sora 平台，确认发送？',
        description: '将已完成的视频任务发布为 Sora 帖子。仅支持 status 为 completed 的视频任务。',
        bodyParams: [
          { name: 'task_id', type: 'string', required: true, description: '已完成的视频任务 ID' },
        ],
        responseExample: `{
  "post_id": "post_xxx"
}`,
      },
      {
        id: 'delete-post',
        method: 'DELETE',
        path: '/v1/posts/:id',
        title: '删除帖子',
        description: '删除已发布的 Sora 帖子。',
        params: [
          { name: 'id', type: 'string', required: true, description: '帖子 ID' },
        ],
        responseExample: `// 204 No Content`,
      },
    ],
  },
  // ── 无水印下载 ──
  {
    id: 'watermark-free',
    title: '无水印下载',
    group: 'tools',
    description: '获取 Sora 视频的无水印下载链接。',
    endpoints: [
      {
        id: 'get-watermark-free',
        method: 'POST',
        path: '/v1/watermark-free',
        title: '获取无水印下载链接',
        description: '传入 Sora 分享链接或视频 ID，返回无水印的源视频下载链接。',
        bodyParams: [
          { name: 'video_id', type: 'string', required: true, description: 'Sora 分享链接或视频 ID（s_xxx 格式）' },
        ],
        responseExample: `{
  "url": "https://..."
}`,
      },
    ],
  },
  // ── 可用模型 ──
  {
    id: 'models',
    title: '可用模型',
    group: 'reference',
    description: '以下为所有支持的模型名称和对应参数。创建视频任务时 model 字段必须使用这些名称。',
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
| sora-2-pro-portrait-hd-25s | 竖屏 | 25s | 1080x1920 |

**视频风格**

在 prompt 中使用 \`{style}\` 标记或通过 \`style\` 参数指定。可用风格：

| 风格 ID | 描述 |
|---------|------|
| anime | 动漫风 |
| retro | 复古风 |
| comic | 漫画风 |
| nostalgic | 怀旧风 |
| golden | 金色调 |
| handheld | 手持拍摄 |
| selfie | 自拍风 |
| news | 新闻风 |
| festive | 节日风 |
| kakalaka | Kakalaka |

**分镜模式**

prompt 中使用 \`[Ns]\` 标记自动进入分镜模式。示例：

\`[5.0s]猫在花园奔跑 [5.0s]猫跳上围墙 [5.0s]猫在阳光下打盹\`

**注意**：Remix 固定使用标准模型，分镜固定使用标准画质。`,
      },
    ],
  },
  // ── 状态码与错误 ──
  {
    id: 'errors',
    title: '状态码与错误',
    group: 'reference',
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
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误（如模型名无效） |
| 401 | 认证失败（API Key 无效或已禁用） |
| 404 | 资源不存在 |
| 500 | 服务内部错误（如 Sora API 调用失败） |
| 503 | 无可用账号 |

**视频/图片任务状态**

| 状态 | 说明 |
|------|------|
| queued | 已提交，等待处理 |
| in_progress | 生成中 |
| completed | 已完成，可下载 |
| failed | 失败 |

**角色状态**

| 状态 | 说明 |
|------|------|
| processing | 处理中（上传 → 轮询 → 定稿） |
| ready | 已就绪，可使用 |
| failed | 处理失败 |

**错误响应格式**

\`{ "error": { "message": "错误描述" } }\``,
      },
    ],
  },
]
