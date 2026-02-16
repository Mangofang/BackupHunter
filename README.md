# FormatCnver
![BackupHunter](https://socialify.git.ci/Mangofang/BackupHunter/image?custom_description=%E8%87%AA%E5%8A%A8%E5%8C%96Web%E5%A4%87%E4%BB%BD%E6%96%87%E4%BB%B6%E6%89%AB%E6%8F%8F%E3%80%81%E7%9B%91%E6%8E%A7%E7%B3%BB%E7%BB%9F&description=1&font=Raleway&forks=1&issues=1&language=1&name=1&owner=1&pattern=Circuit+Board&pulls=1&stargazers=1&theme=Dark)

![BackupHunter](https://img.shields.io/badge/version-v1.0-blue)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8)

> **自动化备份文件扫描、监控系统**

🌐 **[English README](README_EN.md)**

如果你有任何问题或反馈程序问题请提交`Issues`

<img src="https://github.com/Mangofang/BackupHunter/blob/main/img/hander.png" />
<img src="https://github.com/Mangofang/BackupHunter/blob/main/img/hander2.png" />

## 关于：
BackupHunter 是一个基于 `Go` 语言开发的自动化备份文件扫描与监控系统。使用`OneForAll`进行子域名发现，集成备份文件扫描、任务调度等功能，可以帮助安全研究人员自动化地发现和监控目标域名的备份文件。

## 声明：
1. 文中所涉及的技术、思路和工具仅供以安全为目的的学习交流使用，任何人不得将其用于非法用途以及盈利等目的，否则后果自行承担！
2. 水平不高，纯萌新面向Google编程，同时借鉴了很多大佬的代码

## 功能特性
- 📁 **备份文件扫描**
  - 支持自定义扫描字典
  - 并发扫描提高效率
  - 自动下载发现的备份文件

- ⏰ **任务调度**
  - 基于 Cron 的定时任务管理
  - 支持任务的创建、暂停、启动、删除
  - 任务执行状态实时监控

## TODO

> [!TIP]
>
> - [ ... ] OneForAll前端修改 - 正在添加前端对OneForAll的配置修改
> - [ √ ] Docker部署 - 现已经支持使用Docker部署

## 部署
### Docker部署（推荐）
```
curl -o BackupHunter.zip https://github.com/Mangofang/BackupHunter/releases/download/20260217/BackupHunter.zip && unzip BackupHunter.zip
cd BackupHunter
docker-compose up -d --build
```

## 免责声明

1. 本工具仅供安全研究和学习使用
2. 请勿将本工具用于非法用途
3. 使用本工具造成的任何后果由使用者自行承担

## 致谢
[shmilylty](https://github.com/shmilylty): [OneForAll](https://github.com/shmilylty/OneForAll)

[r00tSe7en](https://github.com/r00tSe7en)及[VMsec](https://github.com/VMsec): [ihoneyBakFileScan_Modify](https://github.com/VMsec/ihoneyBakFileScan_Modify)