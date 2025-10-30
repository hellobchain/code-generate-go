package main

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed tpl/*
var defaultFS embed.FS

// 返回实际可用的模板 FS + 是否用户自定义目录
func templateFS() (fsys fs.FS, user bool) {
	if _, err := os.Stat("tpl"); err == nil {
		// 用户侧存在 tpl/ 目录，优先使用
		return os.DirFS("./tpl"), true
	}
	// 回退到 embed
	sub, _ := fs.Sub(defaultFS, "tpl")
	return sub, false
}

// 根据模板名读取内容
func readTemplate(fsys fs.FS, name string) string {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		panic("template " + name + " not found: " + err.Error())
	}
	return string(b)
}
