# 血压管理系统 (Blood Manager)

[![GitHub release](https://img.shields.io/github/v/release/hsieh19/blood-manager)](https://github.com/hsieh19/blood-manager/releases)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/hsieh19/blood-manager)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

一个极简、私密且高效的家庭血压记录与管理系统。专为关注健康的家庭用户设计，支持多用户管理，提供数据备份与还原功能。

## ✨ 主要功能

- **🚀 极简记录**：快速录入收缩压、舒张压和心率，支持备注。
- **📊 历史查询**：支持按日期筛选查询历史记录，数据一目了然。
- **👥 用户管理**：支持管理员创建和管理多个用户账号。
- **💾 数据安全**：
  - 支持本地导出备份（自动生成时间戳文件名）。
  - 支持手动指定路径一键还原数据库。
- **🌓 主题切换**：支持浅色与深色模式，保护视力。
- **🛡️ 安全保障**：支持设置全局自动退出时间（Idle Timeout）。
- **📱 响应式设计**：完美适配手机、平板及桌面端访问。

## 🛠️ 技术栈

- **后端**: Go 1.21+, Gin Web Framework
- **数据库**: SQLite (默认), 同时也支持云端 MySQL 配置
- **前端**: Vanilla JS, Modern CSS (无外部重型框架，极致轻量)
- **认证**: 基于 Session 的身份验证

## 🚀 快速开始

### 本地编译运行

1. 克隆仓库：
   ```bash
   git clone https://github.com/hsieh19/blood-manager.git
   cd blood-manager
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 运行：
   ```bash
   go run main.go
   ```

4. 访问：`http://localhost:8080` (默认管理员账号: admin / admin)

## 📦 版本发布

我们提供预编译的二进制文件，支持以下架构：
- **ARMv7l**: 适用于树莓派等嵌入式设备。

请访问 [Releases](https://github.com/hsieh19/blood-manager/releases) 页面下载。

## ⚖️ 开源协议

MIT License
