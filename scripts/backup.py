#!/usr/bin/env python
#-*-coding:utf8-*-

# @Time	 : 2018/12/07
# @Author: Rock
# @File  : backup.py

import os
import sys
import time
import datetime
import glob
import shutil
import zipfile
import logging
import pprint
import ConfigParser

configFile = os.getcwd()+"/conf/config.ini"
now = time.strftime("%Y%m%d")
sectionDict = {}

# 日志初始化
def logInit(logName):
	logFormat = "%(asctime)s - %(levelname)s - %(message)s"
	logging.basicConfig(filename = logName, level = logging.DEBUG, format = logFormat)
	return logging

# 判断config.init文件是否存在，不存在就退出
def checkConfExist():
	if not os.path.exists(configFile):
		print "Not found config file"
		sys.exit(1)

# 读取配置文件，并以字典的形式保存	
def readConfig():
	conf = ConfigParser.RawConfigParser()
	conf.read(configFile)
	for section in conf.sections():
		section = section.strip()
		#print conf.items(section)     # [('host', '10.155.90.179'), ('target_module', 'backup')]
		sectionDict[section] = {}
		if len(conf.items(section)) > 0:
			for item in conf.items(section):
				key, value = item
				sectionDict[section][key] = value
	#pprint.pprint(sectionDict)

# 检查备份目录是否存在
def checkBackupDirExist(log):
	if sectionDict.get("backup"):
		targetDir = sectionDict.get("backup").get("target_dir").strip()
		if targetDir:
			if not os.path.exists(targetDir):
				os.mkdir(targetDir)
		else:
			targetDir = "/conf_backup"
			if not os.path.exists(targetDir):
				os.mkdir(targetDir)
	else:
		log.error("There is no backup path in the configuration file")
		sys.exit(2)

# 获取本机IP
def getIP():
	ip = os.popen("ifconfig | grep inet | egrep -v '127.0.0.1 | inet6' | grep '10.' | awk '{print $2}' | cut -d':' -f2").read().rstrip("\n")
	return ip

# 获取当天真正备份的目录
def getBackupDir():
	ip = getIP()
        targetDir = sectionDict.get("backup").get("target_dir").strip()
        if not targetDir.endswith("/"):
                targetDir = "%s/" % targetDir
        backupDir = "%s%s/%s" % (targetDir, ip, now)
	return backupDir

# 备份文件和目录
def backupFileAndDir():
	backupDir = getBackupDir()
	if os.path.exists(backupDir) and os.path.isdir(backupDir):
		pass
	else:
		os.makedirs(backupDir)
	if sectionDict.get("file"):
		files = sectionDict.get("file").get("file_list")
		if files:
			if files.find(","):
				fileList = files.split(",")
				for file in fileList:
					if os.path.exists(file) and os.path.isfile(file):
						shutil.copy(file, backupDir)
			else:
				if os.path.exists(file) and os.path.isfile(file):
                                	shutil.copy(file, backupDir)
	if sectionDict.get("directory"):	
		dirs = sectionDict.get("directory").get("dir_list")
		if dirs:
			if dirs.find(","):
				dirList = dirs.split(",")
				for dirName in dirList:
					copyDir(dirName, backupDir)
			else:
				copyDir(dirs, backupDir)
	if sectionDict.get("regexp_file"):
		regexpFiles = sectionDict.get("regexp_file").get("regexp_file_list")
		if regexpFiles:
			if regexpFiles.find(","):
				regexpFileList = regexpFiles.split(",")
				for regexpFile in regexpFileList:
					fileList = glob.glob("%s*" % regexpFile)
					copyFile(fileList, backupDir)	
			else:
				fileList = glob.glob("%s*" % regexpFiles)
				copyFile(fileList, backupDir)

# 拷贝目录
def copyDir(dirName, backupDir):
        if os.path.exists(dirName) and os.path.isdir(dirName):
                try:
                        newBackupDir = "%s/%s" % (backupDir, dirName)
			try:
                        	shutil.copytree(dirName, newBackupDir)
			except shutil.Error as e:
				pass
                except OSError as e:
                        pass

# 拷贝需要正则匹配的文件到需要备份的目录
def copyFile(fileList, backupDir):
	if fileList:
		for file in fileList:
			if os.path.isfile(file):
                        	shutil.copy(file, backupDir)
                        elif os.path.isdir(file):
                        	try:
                        		newBackupDir = "%s/%s" % (backupDir, file)
                                        shutil.copytree(file, newBackupDir)
                                except OSError as e:
                                        pass

# 打包拷贝的文件和目录
def tarBackup(log):
	zipDir = getBackupDir()
	storeZipDir = os.path.dirname(zipDir)
	zipName = "%s/%s.zip" % (storeZipDir, now)
	if os.path.exists(zipName):
		os.remove(zipName)
	f = zipfile.ZipFile(zipName, 'w', zipfile.ZIP_DEFLATED)
	try:
		for parent, dirnames, filenames in os.walk(zipDir):
        		for filename in filenames:
                		zipFile = os.path.join(parent, filename)
                		f.write(zipFile)
	except Exception as e:
		log.error("zip file error:", e)
		pass
	else:
		shutil.rmtree(zipDir)
	finally:
		f.close()
		
# 定期清理打包文件
def deleteBackup(log):
	saveDay = int(sectionDict.get("backup").get("save_days"))
	if isinstance(saveDay, int) and saveDay > 0:
		pass
	else:
		saveDay = 3
	needStoreBackups = []
	now = datetime.datetime.now()
	for i in range(0, saveDay):
		beforeDay = now - datetime.timedelta(days=i)
		newBeforeDay = beforeDay.strftime('%Y%m%d')
		storeBackupName = "%s.zip" % str(newBeforeDay)
		needStoreBackups.append(storeBackupName)
	backupDir = getBackupDir()
	newBackupDir = os.path.dirname(backupDir)
	for storeBackupName in glob.glob(newBackupDir+"/*.zip"):
		storeBackupNameSuffix = os.path.basename(storeBackupName)
		if storeBackupNameSuffix not in needStoreBackups:
			log.info("%s has been deleted" % storeBackupNameSuffix)
			os.remove(storeBackupName)

# 同步打包文件到备份服务器
def rsyncTarFile(log):
	remoteHost = sectionDict.get("rsync").get("host")
	targetModule = sectionDict.get("rsync").get("target_module")
	rsyncPasswd = sectionDict.get("rsync").get("rsync_passwd")
	rsyncDir = os.path.dirname(getBackupDir())
	if remoteHost and targetModule and rsyncPasswd:
		rsyncCmd = "rsync -a --password-file=%s %s root@%s::%s" % (rsyncPasswd, rsyncDir, remoteHost, targetModule)
		out = os.system(rsyncCmd)
		if int(out) != 0:
			log.error("rsync error")

# 日志切割
def cutLog():
	logPath = sectionDict.get("log").get("log_path")
	cutLogSize = sectionDict.get("log").get("cut_size")
	if not cutLogSize or not isinstance(cutLogSize, int):
		cutLogSize = 1024
	if logPath and os.path.exists(logPath):
		logSize = (os.path.getsize(logPath))/1024
		if logSize > cutLogSize:
			prefixLogName, logNameSuffix = os.path.splitext(os.path.basename(logPath))
			newLogName = "%s_%s.log" % (prefixLogName, now) 
			newLogPath = "%s/%s" % (os.path.dirname(logPath), newLogName)
			shutil.move(logPath, newLogPath)

# 汇总函数
def main():
	checkConfExist()
	readConfig()
	if sectionDict.get("log").get("log_path"):
		logPath = sectionDict.get("log").get("log_path")
	else:
		logPath = "/var/log/backup.log"
	log = logInit(logPath)
	checkBackupDirExist(log)
	backupFileAndDir()
	tarBackup(log)
	rsyncTarFile(log)
	deleteBackup(log)
	cutLog()

if __name__ == "__main__":
	main()
