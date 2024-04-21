package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Print the command prompt with relevant info.
func printCommandLine() {
	user, err := user.Current()
	if nil != err {
		panic("Error getting user information!")
	}

	//hostname, err := os.Hostname()
	//if nil != err {
	//	panic("Error getting hostname!")
	//}

	cwd, err := os.Getwd()
	if nil != err {
		panic("Error getting current working directory!")
	}

	fmt.Printf("\x1b[34;1m%s\x1b[0m \x1b[36;1m%s-> \x1b[0m", user.Username, cwd)
}

// Print working directory.
func pwd() {
	cwd, err := os.Getwd()
	if nil != err {
		fmt.Println("Error getting current working directory!")
		return
	}
	fmt.Println(cwd)
}
func clearScreen() {
	// Use the "clear" command on Unix-like systems
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// User information lookup.
func finger(args []string) {
	var usr *user.User
	var err error

	if 1 == len(args) {
		usr, err = user.Current()
		if nil != err {
			fmt.Printf("\x1b[31;1mfinger: failed to get current user.\x1b[0m\n")
			return
		}
	} else {
		usr, err = user.Lookup(args[1])
		if nil != err {
			fmt.Printf("\x1b[31;1mfinger: %s: no such user.\x1b[0m\n", args[1])
			return
		}
	}

	fmt.Printf("\x1b[32;1mLogin:\x1b[0m %s \t%-25s\x1b[32;1mName:\x1b[0m %s \n", usr.Username, "", usr.Name)
	fmt.Printf("\x1b[32;1mUID:\x1b[0m %s \t%-25s\x1b[32;1mGID:\x1b[0m %s \n", usr.Uid, "", usr.Gid)
	fmt.Printf("\x1b[32;1mHome:\x1b[0m %s\n", usr.HomeDir)
}

// Prints environment variables.
func env() {
	env := os.Environ()
	for i := 0; i < len(env); i++ {
		fmt.Println(env[i])
	}
}

// Changes the current working directory.
func changeDir(args []string) {
	if len(args) == 1 {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("\x1b[31;1mfinger: failed to get current user.\x1b[0m\n")
			return
		}
		if err := os.Chdir(usr.HomeDir); err != nil {
			fmt.Printf("\x1b[31;1mcd: %s: No such file or directory\x1b[0m\n", usr.HomeDir)
		}
	} else if args[1] == "~" {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("\x1b[31;1mfinger: failed to get current user.\x1b[0m\n")
			return
		}
		if err := os.Chdir(usr.HomeDir); err != nil {
			fmt.Printf("\x1b[31;1mcd: %s: No such file or directory\x1b[0m\n", usr.HomeDir)
		}
	} else if err := os.Chdir(args[1]); err != nil {
		fmt.Printf("\x1b[31;1mcd: %s: No such file or directory\x1b[0m\n", args[1])
	}
}

// List directory contents
func ll(args []string) {
	var name string
	var err error

	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("\x1b[31;1mls: Can't get current working directory\x1b[0m\n")
		return
	}

	if 1 == len(args) {
		name = path + "/."
	} else {
		if strings.HasPrefix(args[1], "/") {
			name = args[1]
		} else {
			name = path + "/" + args[1]
		}
	}

	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	mode := file.Mode()

	if mode.IsDir() {
		llDir(name)
	} else {
		llFile(name)
	}
}

// Prints information about a directory.
func llDir(name string) {
	llFile(name + "/.")
	llFile(name + "/..")

	dir, err := os.Open(name)

	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't open %s\x1b[0m\n", name)
		return
	}

	dirs, err := dir.Readdirnames(0)

	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't read entries in %s\x1b[0m\n", name)
		return
	}

	for i := 0; i < len(dirs); i++ {
		llFile(name + "/" + dirs[i])
	}
}

// Prints information about a file.
func llFile(name string) {
	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	perms := file.Mode().String()

	if strings.HasPrefix(perms, "L") {
		perms = perms[1:len(perms)]
		perms = "l" + perms
	}

	fmt.Printf("%s ", perms)

	var stNumLinks uint64
	sys := file.Sys()
	if sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			stNumLinks = uint64(stat.Nlink)
		}
	}

	fmt.Printf("%d\t", stNumLinks)
	fmt.Printf("%d\t", sys.(*syscall.Stat_t).Uid)
	fmt.Printf("%d\t", sys.(*syscall.Stat_t).Gid)
	fmt.Printf("%d\t", file.Size())
	fmt.Printf("%s ", file.ModTime().Format(time.UnixDate))
	fmt.Printf("%s", file.Name())

	if 0 != (file.Mode() & os.ModeSymlink) {
		linkname, err := os.Readlink(file.Name())
		if err != nil {
			fmt.Printf("\n")
			return
		}
		fmt.Printf(" -> %s\n", linkname)
	} else {
		fmt.Printf("\n")
	}
}

// List directory contents in compact form.
func ls(args []string) {
	var name string
	var err error

	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("\x1b[31;1mls: Can't get current working directory\x1b[0m\n")
		return
	}

	if 1 == len(args) {
		name = path + "/."
	} else {
		if strings.HasPrefix(args[1], "/") {
			name = args[1]
		} else {
			name = path + "/" + args[1]
		}
	}

	file, err := os.Lstat(name)
	if nil != err {
		fmt.Printf("\x1b[31;1mls: Can't stat %s\x1b[0m\n", name)
		return
	}

	mode := file.Mode()

	if mode.IsDir() {
		dir, err := os.Open(name)
		if nil != err {
			fmt.Printf("\x1b[31;1mls: Can't open %s\x1b[0m\n", name)
			return
		}

		dirs, err := dir.Readdirnames(0)
		if nil != err {
			fmt.Printf("\x1b[31;1mls: Can't read entries in %s\x1b[0m\n", name)
			return
		}

		for i := 0; i < len(dirs); i++ {
			if strings.HasPrefix(dirs[i], ".") {
				continue
			}
			fmt.Printf("%s\n", dirs[i])
		}
		if 0 != len(dirs) {
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("%s\n", name)
	}
}

func printOut(line string) {
	fmt.Printf("%s", line)
}

// Parse the string the user enters in to the command prompt.
func parseCommand(line string) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		for {
			<-s
		}
	}()
	cmd := exec.Command("/bin/sh", "-c", line)
	cmd.Stdout = os.Stdout //&out
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func listEnvironmentVariables() {
	env := os.Environ()
	page := 1
	for i, v := range env {
		if i%20 == 0 && i != 0 {
			fmt.Print("--More--")
			fmt.Scanln()
			page++
			fmt.Println(strings.Repeat("-", 80))
		}
		fmt.Println(v)
	}
}

func help() {
	fmt.Println("cd <directory>   - Change the current default directory to <directory>.")
	fmt.Println("                   If the directory does not exist, an appropriate error should be reported.")
	fmt.Println("                   This command also changes the PWD environment variable.")
	fmt.Println("clr              - Clear the screen.")
	fmt.Println("dir <directory>  - List the contents of directory <directory>.")
	fmt.Println("environ          - List all the environment strings.")
	fmt.Println("echo <comment>   - Display <comment> on the display followed by a new line.")
	fmt.Println("                   Multiple spaces/tabs may be reduced to a single space.")
	fmt.Println("help             - Display the user manual. ")
	fmt.Println("ll               - List detailed information about files and directories in the current directory.")
	fmt.Println("pause            - Pause operation of the shell until 'Enter' is pressed.")
	fmt.Println("quit             - Quit the shell.")
}

func echo(args []string) {
	comment := strings.Join(args[1:], " ")

	re := regexp.MustCompile(`[\s\t]+`)
	comment = re.ReplaceAllString(comment, " ")

	fmt.Println(comment)
}
func pause() {
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
func specialCommand() {
	asciiArt := "\n" +
		"                 .88888888:.\n" +
		"                88888888.88888.\n" +
		"              .8888888888888888.\n" +
		"              888888888888888888\n" +
		"              88' _`88'_  `88888\n" +
		"              88 88 88 88  88888\n" +
		"              88_88_::_88_:88888\n" +
		"              88:::,::,:::::8888\n" +
		"              88`:::::::::'`8888\n" +
		"             .88  `::::'    8:88.\n" +
		"            8888            `8:888.\n" +
		"          .8888'             `888888.\n" +
		"         .8888:..  .::.  ...:'8888888:.\n" +
		"        .8888.'     :'     `'::`88:88888\n" +
		"       .8888        '         `.888:8888.\n" +
		"      888:8         .           888:88888\n" +
		"    .888:88        .:           888:88888:\n" +
		"    8888888.       ::           88:888888\n" +
		"    `.::.888.      ::          .88888888\n" +
		"   .::::::.888.    ::         :::`8888'.:.\n" +
		"  ::::::::::.888   '         .::::::::::::\n" +
		"  ::::::::::::.8    '      .:8::::::::::::.\n" +
		" .::::::::::::::.        .:888:::::::::::::\n" +
		" :::::::::::::::88:.__..:88888:::::::::::'\n" +
		"  `'.:::::::::::88888888888.88:::::::::'\n" +
		"        `':::_:' -- '' -'-' `':_::::'\n"

	fmt.Println(asciiArt)
}

func main() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		for {
			<-s
			fmt.Printf("\n")
		}

	}()
	clearScreen()
	asciiArt :=
		"\x1b[38;5;141m  ____          ____   _            _  _  \n" +
			" / ___|  ___   / ___| | |__    ___ | || | \n" +
			"| |  _  / _ \\  \\___ \\ | '_ \\  / _ \\| || | \n" +
			"| |_| || (_) |  ___) || | | ||  __/| || | \n" +
			" \\____| \\___/  |____/ |_| |_| \\___||_||_| \n" +
			"             ____                         \n" +
			"            | __ )  _   _                 \n" +
			"            |  _ \\ | | | |                \n" +
			"            | |_) || |_| |                \n" +
			"            |____/  \\__, |                \n" +
			"         _          |___/                 \n" +
			"        | | _   _   __ _  _ __            \n" +
			"     _  | || | | | / _` || '_ \\           \n" +
			"    | |_| || |_| || (_| || | | |          \n" +
			"     \\___/  \\__,_| \\__,_||_| |_|          \n" +
			"                ___                       \n" +
			"               ( _ )                      \n" +
			"               / _ \\/\\                    \n" +
			"              | (_>  <                    \n" +
			"    ____       \\___/\\/     _  _           \n" +
			"   / ___| __ _  _ __ ___  (_)| |  __ _    \n" +
			"  | |    / _` || '_ ` _ \\ | || |/ _` |   \n" +
			"  | |___| (_| || | | | | || || | (_| |   \n" +
			"   \\____|\\__,_||_| |_| |_||_||_| \\__,_|   \n\x1b[0m"
	fmt.Println(asciiArt)

	fmt.Println("It supports the following commands: dir, clr, environ, echo, pause, help, cd, ll, ls, pwd.")

	printCommandLine()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		line := scanner.Text()
		args := strings.Split(line, " ")

		if "" == args[0] || " " == args[0] {
			continue
		} else if "help" == args[0] {
			help()
		} else if "dir" == args[0] {
			ls(args)
		} else if "ll" == args[0] {
			ll(args)
		} else if "pwd" == args[0] {
			pwd()
		} else if "cd" == args[0] {
			changeDir(args)
		} else if "clr" == args[0] {
			clearScreen()
		} else if "environ" == args[0] {
			listEnvironmentVariables()
		} else if "echo" == args[0] {
			echo(args)
		} else if "pause" == args[0] {
			pause()
		} else if "special" == args[0] {
			specialCommand()
		} else if "quit" == args[0] {
			os.Exit(0)
		} else {
			parseCommand(line)
		}
		printCommandLine()
	}
}
