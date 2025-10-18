package main

import "fmt"

type ParsedArgs struct {
	Command    Command
	JdkVersion string
	MvnVersion string
	ListType   string
}

type Command int

const (
	CmdNone Command = iota
	CmdUseJdk
	CmdUseMvn
	CmdList
	CmdInstallJdk
	CmdInstallMvn
	CmdUninstallJdk
	CmdUninstallMvn
	CmdHelp
)

func (c Command) String() string {
	switch c {
	case CmdUseJdk:
		return "use-jdk"
	case CmdUseMvn:
		return "use-mvn"
	case CmdList:
		return "list"
	case CmdInstallJdk:
		return "install-jdk"
	case CmdInstallMvn:
		return "install-mvn"
	case CmdUninstallJdk:
		return "uninstall-jdk"
	case CmdUninstallMvn:
		return "uninstall-mvn"
	case CmdHelp:
		return "help"
	default:
		return "none"
	}
}
func parseArgs(args []string) (ParsedArgs, error) {
	var parsed ParsedArgs

	if len(args) == 0 {
		parsed.Command = CmdHelp
		return parsed, nil
	}

	switch args[0] {
	case "use-jdk":
		parsed.Command = CmdUseJdk
		if len(args) > 1 {
			parsed.JdkVersion = args[1]
		}
	case "use-mvn":
		parsed.Command = CmdUseMvn
		if len(args) > 1 {
			parsed.MvnVersion = args[1]
		}
	case "list":
		parsed.Command = CmdList
		if len(args) > 1 {
			parsed.ListType = args[1]
		}
	case "install-jdk":
		parsed.Command = CmdInstallJdk
		if len(args) > 1 {
			parsed.JdkVersion = args[1]
		}
	case "install-mvn":
		parsed.Command = CmdInstallMvn
		if len(args) > 1 {
			parsed.MvnVersion = args[1]
		}
	case "uninstall-jdk":
		parsed.Command = CmdUninstallJdk
		if len(args) > 1 {
			parsed.JdkVersion = args[1]
		}
	case "uninstall-mvn":
		parsed.Command = CmdUninstallMvn
		if len(args) > 1 {
			parsed.MvnVersion = args[1]
		}
	case "help":
		parsed.Command = CmdHelp
	default:
		return parsed, fmt.Errorf("unknown argument '%s'", args[0])
	}

	return parsed, nil
}
