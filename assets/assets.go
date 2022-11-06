package assets

import "embed"

//go:embed client/* server/*
var Assets embed.FS
