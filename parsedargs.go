package main

type ParsedArgs struct {
	JdkVersion          string
	MvnVersion          string
	ListCommand         string
	UseJdkCommand       bool
	UseMvnCommand       bool
	InstallJdkCommand   bool
	InstallMvnCommand   bool
	UninstallJdkCommand bool
	UninstallMvnCommand bool
	HelpCommand         bool
}
