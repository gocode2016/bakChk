# 备份检查系统

**备份检查系统主要包括备份脚本和备份检查程序两部分组成，现详细说明如下:**

* 备份脚本

备份脚本（scripts目录）主要是备份服务器中一些重要的文件或者目录，由于备份的文件或者目录众多，且各式各样，所以提供了配置文件定义的方式，方便扩展和一些定制
化的备份策略，还支持备份路径，日志记录，日志切割，备份保留策略和远程备份的地址选择等等动态配置的自定义选项，远程同步使用的rsync的方式，所以远程备份机需要
配置好相关同步模块，同步的用户名和密码等等，这些在配置文件中也做成了动态自定义选项。

* 备份检查程序

首先说明的是备份检查程序还没有经过严格的并发测试，可能还存在一些线程安全的问题，话不多说，上图，一目了然：
![](https://github.com/cash666/bakChk/blob/master/registerServer/screenShots/%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E5%90%8C%E6%AD%A5%E4%B8%8E%E6%A3%80%E6%9F%A5%E6%9E%B6%E6%9E%84%E5%9B%BE.png)
