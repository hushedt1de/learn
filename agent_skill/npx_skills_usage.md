# npx skills 使用指南

`skills` CLI 用于为 Codex、Claude Code 等编码 Agent 安装和管理 skills。

## 1. 安装位置

### 安装到当前项目

默认情况下，skill 会安装到当前项目：

```bash
npx skills@latest add mattpocock/skills
```

常见的相关文件是：

```text
.agents/skills/   # skill 文件
skills-lock.json  # skill 来源、路径和版本哈希等安装记录
```

这种方式适合只在当前项目中使用的 skill。

### 全局安装

希望多个项目都能使用时，加上 `--global` 或 `-g`：

```bash
npx skills@latest add mattpocock/skills --global
```

## 2. 安装指定的 skill

安装仓库中的部分 skills，使用 `--skill` 或 `-s`：

```bash
npx skills@latest add mattpocock/skills \
  --skill ask-matt grill-with-docs to-prd to-issues
```

指定安装到 Codex，并跳过确认提示：

```bash
npx skills@latest add mattpocock/skills \
  --agent codex \
  --skill ask-matt grill-with-docs to-prd to-issues \
  -y
```

不指定 `--skill` 时，CLI 可能显示交互选择界面；在某些环境中也可能自动安装仓库内的全部 skills。因此，需要精确控制安装范围时，应明确使用 `--skill`。

## 3. 查看已安装的 skills

查看当前项目：

```bash
npx skills@latest list
```

查看当前项目的 JSON 格式结果：

```bash
npx skills@latest list --json
```

查看全局 skills：

```bash
npx skills@latest list --global
```

## 4. 删除指定的 skill

删除当前项目中的指定 skills：

```bash
npx skills@latest remove ask-matt grill-me -y
```

也可以使用 `--skill`：

```bash
npx skills@latest remove --skill ask-matt grill-me -y
```

删除全局的指定 skills：

```bash
npx skills@latest remove ask-matt grill-me --global -y
```

去掉 `-y` 后，CLI 会要求确认。

## 5. 删除全部 skills

删除当前项目作用域中的全部 skills：

```bash
npx skills@latest remove --all -y
```

删除全局作用域中的全部 skills：

```bash
npx skills@latest remove --all --global -y
```

> 注意：`--all` 会删除对应作用域中的所有 skills。

## 6. 手动删除时的注意事项

直接删除 `.agents/skills/` 会让其中的 skills 不再可用，但可能留下 `skills-lock.json` 中的安装记录。

优先使用 `npx skills@latest remove`。如果确认项目中不再保留任何 skill，可以在卸载后检查并清理：

```text
.agents/skills/
skills-lock.json
```

不要直接删除整个 `.agents/`，因为其中可能还有其他 Agent 配置。

## 7. 常用命令速查

| 目的 | 命令 |
| --- | --- |
| 安装到当前项目 | `npx skills@latest add mattpocock/skills` |
| 安装指定 skills | `npx skills@latest add mattpocock/skills --skill ask-matt grill-me` |
| 全局安装 | `npx skills@latest add mattpocock/skills --global` |
| 查看项目 skills | `npx skills@latest list` |
| 查看全局 skills | `npx skills@latest list --global` |
| 删除指定 skills | `npx skills@latest remove ask-matt grill-me -y` |
| 删除项目全部 skills | `npx skills@latest remove --all -y` |
| 删除全局全部 skills | `npx skills@latest remove --all --global -y` |

