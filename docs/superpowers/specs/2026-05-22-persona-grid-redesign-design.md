# Persona 管理页面重构设计

## 概述

将现有的 `PersonaManagePage` 从长方形卡片列表重构为**正方形网格卡片**，并新增基于 **archetype（思维原型）** 的筛选功能。

## 动机

- 现有卡片为长方形，信息密度不均，视觉上不够聚焦
- 缺少按人物"类型"筛选的能力，用户难以快速定位感兴趣的人物
- 单一职业标签（如"产品经理"）无法描述多面人物（如 Steve Jobs 同时是设计师、创始人和愿景家）

## 数据模型变更

### 新增字段：archetype

在 `Persona` schema 中新增 `archetype` 字段：

```typescript
// frontend/src/types/persona.ts
interface Persona {
  // ... existing fields
  archetype: string; // 新增
}

// backend/internal/persona/persona.go
type Persona struct {
  // ... existing fields
  Archetype string `json:"archetype"` // 新增
}
```

**Archetype 取值范围（5 个）：**

| Archetype | 中文名 | 代表人物 |
|-----------|--------|----------|
| Visionary | 远见者 | Steve Jobs |
| Engineer | 工程师 | Elon Musk |
| Philosopher | 哲思者 | Naval Ravikant |
| Craftsman | 匠人 | 张小龙 |
| Operator | 运营者 | 张一鸣 |

**影响范围：**
- 5 个现有 persona JSON 文件需要补充 `archetype` 字段
- `PersonaSummary` 类型需要暴露 `archetype`（供列表页使用）
- PersonaEditor 需要增加 archetype 选择/输入

### PersonaSummary 扩展

```typescript
// frontend/src/types/index.ts
export interface PersonaSummary {
  id: string;
  name: string;
  display_name: string;
  avatar: string;
  role_title: string;
  description: string;
  tags: string[];
  archetype: string; // 新增
}
```

Backend `ListPersonas` handler 已在响应中包含所有字段，只需确保前端类型匹配。

## UI 设计

### 筛选栏

- 位置：页面标题下方，网格上方
- 样式：pill 形状标签，横向排列，可换行
- 交互：单选，点击后高亮，网格实时过滤
- 默认态："全部" 高亮，显示所有人物
- 选项：全部 | Visionary | Engineer | Philosopher | Craftsman | Operator
- 多语言：archetype 名称走 i18n（英文显示原词，中文显示对应翻译）

### 人物网格卡片

**默认态：**
- 形状：正方形（`aspect-ratio: 1 / 1`）
- 布局：flex 垂直居中
- 内容：
  - 头像：64×64px，圆角 12px（`rounded-xl`）
    - 若使用真实图片：`<img>` + `object-fit: cover`
    - 若使用 emoji：40px 字号
  - 名字：14px，font-semibold，text-black/95
  - archetype 标签：11px，text-[#a39e98]
- 背景：白色，圆角 12px，细边框（`border-black/[0.06]`）
- 悬停：轻微阴影（`shadow-[rgba(0,0,0,0.04)_0px_4px_18px]`）

**Hover 覆盖层：**
- 触发：鼠标悬停在卡片上
- 效果：半透明白色背景（`bg-white/98`）覆盖整张卡片，淡入（`opacity` 过渡 0.2s）
- 内容：
  - description：12px，最多 5 行（`line-clamp-5`），居中
  - 操作按钮：编辑（蓝色）+ 删除（红色），11px pill 按钮
- 注意：hover 层本身需要处理鼠标事件（编辑/删除点击）

**Grid 布局：**
- `grid-template-columns: repeat(auto-fill, minmax(160px, 1fr))`
- gap: 12px
- 自适应列数，小屏自动减少列数

### 空状态

筛选无结果时显示：
- 图标：🔍
- 文案："没有符合条件的人物" / "No personas match this filter"

## 头像方案

支持两种头像形式，共存：

1. **真实图片**：`<img src={avatar}>`，64×64px，`rounded-xl`，`object-fit: cover`
2. **Emoji**：`<span>`，40px 字号

判断方式：若 `avatar` 值以 `http` 开头或包含 `/`，按图片处理；否则按 emoji 渲染。

当前仅 Steve Jobs 使用真实图片（像素风头像），其余保持 emoji。

## i18n

新增翻译键：

```typescript
// archetype 名称
archetypeVisionary:    { 'zh-CN': '远见者', 'en-US': 'Visionary' },
archetypeEngineer:     { 'zh-CN': '工程师', 'en-US': 'Engineer' },
archetypePhilosopher:  { 'zh-CN': '哲思者', 'en-US': 'Philosopher' },
archetypeCraftsman:    { 'zh-CN': '匠人',   'en-US': 'Craftsman' },
archetypeOperator:     { 'zh-CN': '运营者', 'en-US': 'Operator' },

// 筛选栏
labelFilterAll:        { 'zh-CN': '全部',   'en-US': 'All' },
msgNoFilterResults:    { 'zh-CN': '没有符合条件的人物', 'en-US': 'No personas match this filter' },
```

Archetype 显示逻辑：
- 后端存储和传输使用英文标识（"Visionary", "Engineer"...）
- 前端通过 `t('archetype' + archetype)` 映射为本地化名称
- 若 archetype 为空或未匹配，不显示标签

## 实现要点

### 组件拆分建议

将 `PersonaManagePage` 拆分为：
- `PersonaManagePage`：页面布局、数据获取、筛选状态管理
- `ArchetypeFilter`：筛选栏组件，接收 archetype 列表和当前选中项
- `PersonaGrid`：网格容器
- `PersonaCard`：单张卡片（默认态 + hover 态）

### 状态管理

```typescript
// PersonaManagePage
const [filterArchetype, setFilterArchetype] = useState<string | null>(null);

const filteredPersonas = useMemo(() => {
  if (!filterArchetype) return personas;
  return personas.filter(p => p.archetype === filterArchetype);
}, [personas, filterArchetype]);
```

### 性能考虑

- 使用 `useMemo` 缓存过滤结果
- 卡片使用 CSS `transition` 实现 hover 效果，避免 React 重渲染
- hover 覆盖层使用 `pointer-events` 控制，确保点击穿透正确

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `backend/internal/persona/persona.go` | 新增 `Archetype string` 字段 |
| `backend/internal/api/router.go` | `PersonaSummary` 新增 `Archetype` 字段 |
| `frontend/src/types/persona.ts` | 新增 `archetype: string` |
| `frontend/src/types/index.ts` | `PersonaSummary` 新增 `archetype` |
| `frontend/src/i18n/translations.ts` | 新增 archetype 相关翻译键 |
| `frontend/src/components/PersonaCard.tsx` | **新增** 人物卡片组件 |
| `frontend/src/components/ArchetypeFilter.tsx` | **新增** 筛选栏组件 |
| `frontend/src/pages/PersonaManagePage.tsx` | 重构为网格布局 + 筛选 |
| `frontend/src/components/PersonaEditor.tsx` | 新增 archetype 输入/选择 |
| `backend/personas/*.json` | 补充 `archetype` 字段（5 个文件） |
| `frontend/public/personas/` | **可选** 存放头像图片资源 |

## 风险与回退

- 若头像图片加载失败，fallback 为显示首字母或默认 emoji
- archetype 字段在旧数据中可能缺失，UI 需要处理空值（不显示标签）
- 头像图片路径需考虑构建后的静态资源引用方式（Vite `public/` 目录）
