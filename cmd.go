package main

import (
	"bufio"
	"fmt"
	"os"

	alpm "github.com/jguer/go-alpm"
)

var cmdArgs = makeArguments()

func usage() {
	fmt.Println(`Usage:
    yay
    yay <operation> [...]
    yay <package(s)>

operations:
    yay {-h --help}
    yay {-V --version}
    yay {-D --database}    <options> <package(s)>
    yay {-F --files}       [options] [package(s)]
    yay {-Q --query}       [options] [package(s)]
    yay {-R --remove}      [options] <package(s)>
    yay {-S --sync}        [options] [package(s)]
    yay {-T --deptest}     [options] [package(s)]
    yay {-U --upgrade}     [options] <file(s)>

New operations:
    yay {-Y --yay}         [options] [package(s)]
    yay {-P --show}        [options]
    yay {-G --getpkgbuild} [package(s)]

New options:
       --repo             Assume targets are from the repositories
    -a --aur              Assume targets are from the AUR

Permanent configuration options:
    --save                Causes the following options to be saved back to the
                          config file when used

    --aururl      <url>   Set an alternative AUR URL
    --builddir    <dir>   Directory used to download and run PKBUILDS
    --editor      <file>  Editor to use when editing PKGBUILDs
    --editorflags <flags> Pass arguments to editor
    --makepkg     <file>  makepkg command to use
    --mflags      <flags> Pass arguments to makepkg
    --pacman      <file>  pacman command to use
    --tar         <file>  bsdtar command to use
    --git         <file>  git command to use
    --gitflags    <flags> Pass arguments to git
    --gpg         <file>  gpg command to use
    --gpgflags    <flags> Pass arguments to gpg
    --config      <file>  pacman.conf file to use
    --makepkgconf <file>  makepkg.conf file to use
    --nomakepkgconf       Use the default makepkg.conf

    --requestsplitn <n>   Max amount of packages to query per AUR request
    --completioninterval  <n> Time in days to to refresh completion cache
    --sortby    <field>   Sort AUR results by a specific field during search
    --answerclean   <a>   Set a predetermined answer for the clean build menu
    --answerdiff    <a>   Set a predetermined answer for the diff menu
    --answeredit    <a>   Set a predetermined answer for the edit pkgbuild menu
    --answerupgrade <a>   Set a predetermined answer for the upgrade menu
    --noanswerclean       Unset the answer for the clean build menu
    --noanswerdiff        Unset the answer for the edit diff menu
    --noansweredit        Unset the answer for the edit pkgbuild menu
    --noanswerupgrade     Unset the answer for the upgrade menu
    --cleanmenu           Give the option to clean build PKGBUILDS
    --diffmenu            Give the option to show diffs for build files
    --editmenu            Give the option to edit/view PKGBUILDS
    --upgrademenu         Show a detailed list of updates with the option to skip any
    --nocleanmenu         Don't clean build PKGBUILDS
    --nodiffmenu          Don't show diffs for build files
    --noeditmenu          Don't edit/view PKGBUILDS
    --noupgrademenu       Don't show the upgrade menu
    --askremovemake       Ask to remove makedepends after install
    --removemake          Remove makedepends after install
    --noremovemake        Don't remove makedepends after install

    --cleanafter          Remove package sources after successful install
    --nocleanafter        Do not remove package sources after successful build
    --bottomup            Shows AUR's packages first and then repository's
    --topdown             Shows repository's packages first and then AUR's

    --devel               Check development packages during sysupgrade
    --nodevel             Do not check development packages
    --gitclone            Use git clone for PKGBUILD retrieval
    --nogitclone          Never use git clone for PKGBUILD retrieval
    --rebuild             Always build target packages
    --rebuildall          Always build all AUR packages
    --norebuild           Skip package build if in cache and up to date
    --rebuildtree         Always build all AUR packages even if installed
    --redownload          Always download pkgbuilds of targets
    --noredownload        Skip pkgbuild download if in cache and up to date
    --redownloadall       Always download pkgbuilds of all AUR packages
    --provides            Look for matching provders when searching for packages
    --noprovides          Just look for packages by pkgname
    --pgpfetch            Prompt to import PGP keys from PKGBUILDs
    --nopgpfetch          Don't prompt to import PGP keys
    --useask              Automatically resolve conflicts using pacman's ask flag
    --nouseask            Confirm conflicts manually during the install
    --combinedupgrade     Refresh then perform the repo and AUR upgrade together
    --nocombinedupgrade   Perform the repo upgrade and AUR upgrade separately

    --sudoloop            Loop sudo calls in the background to avoid timeout
    --nosudoloop          Do not loop sudo calls in the background

    --timeupdate          Check packages' AUR page for changes during sysupgrade
    --notimeupdate        Do not check packages' AUR page for changes

show specific options:
    -c --complete         Used for completions
    -d --defaultconfig    Print default yay configuration
    -g --currentconfig    Print current yay configuration
    -s --stats            Display system package statistics
    -w --news             Print arch news

yay specific options:
    -c --clean            Remove unneeded dependencies
       --gendb            Generates development package DB used for updating

getpkgbuild specific options:
    -f --force            Force download for existing tar packages

If no arguments are provided 'yay -Syu' will be performed
If no operation is provided -Y will be assumed`)
}

func handleCmd() (err error) {
	if cmdArgs.existsArg("h", "help") {
		err = handleHelp()
		return
	}

	if config.boolean["sudoloop"] && cmdArgs.needRoot() {
		sudoLoopBackground()
	}

	switch cmdArgs.op {
	case "V", "version":
		handleVersion()
	case "D", "database":
		err = show(passToPacman(cmdArgs))
	case "F", "files":
		err = show(passToPacman(cmdArgs))
	case "Q", "query":
		err = handleQuery()
	case "R", "remove":
		err = handleRemove()
	case "S", "sync":
		err = handleSync()
	case "T", "deptest":
		err = show(passToPacman(cmdArgs))
	case "U", "upgrade":
		err = show(passToPacman(cmdArgs))
	case "G", "getpkgbuild":
		err = handleGetpkgbuild()
	case "P", "show":
		err = handlePrint()
	case "Y", "--yay":
		err = handleYay()
	default:
		//this means we allowed an op but not implement it
		//if this happens it an error in the code and not the usage
		err = fmt.Errorf("unhandled operation")
	}

	return
}

func handleQuery() error {
	if cmdArgs.existsArg("u", "upgrades") {
		return printUpdateList(cmdArgs)
	}
	return show(passToPacman(cmdArgs))
}

func handleHelp() error {
	if cmdArgs.op == "Y" || cmdArgs.op == "yay" {
		usage()
		return nil
	}
	return show(passToPacman(cmdArgs))
}

func handleVersion() {
	fmt.Printf("yay v%s - libalpm v%s\n", version, alpm.Version())
}

func handlePrint() (err error) {
	switch {
	case cmdArgs.existsArg("d", "defaultconfig"):
		var tmpConfig yayConfig
		tmpConfig.defaultSettings()
		tmpConfig.expandEnv()
		fmt.Printf("%v", tmpConfig)
	case cmdArgs.existsArg("g", "currentconfig"):
		fmt.Printf("%v", config)
	case cmdArgs.existsArg("n", "numberupgrades"):
		err = printNumberOfUpdates()
	case cmdArgs.existsArg("u", "upgrades"):
		err = printUpdateList(cmdArgs)
	case cmdArgs.existsArg("w", "news"):
		err = printNewsFeed()
	case cmdArgs.existsDouble("c", "complete"):
		complete(true)
	case cmdArgs.existsArg("c", "complete"):
		complete(false)
	case cmdArgs.existsArg("s", "stats"):
		err = localStatistics()
	default:
		err = nil
	}
	return err
}

func handleYay() error {
	//_, options, targets := cmdArgs.formatArgs()
	if cmdArgs.existsArg("gendb") {
		return createDevelDB()
	}
	if cmdArgs.existsDouble("c") {
		return cleanDependencies(true)
	}
	if cmdArgs.existsArg("c", "clean") {
		return cleanDependencies(false)
	}
	if len(cmdArgs.targets) > 0 {
		return handleYogurt()
	}
	return nil
}

func handleGetpkgbuild() error {
	return getPkgbuilds(cmdArgs.targets)
}

func handleYogurt() error {
	config.searchMode = numberMenu
	return displayNumberMenu(cmdArgs.targets)
}

func handleSync() error {
	targets := cmdArgs.targets

	if cmdArgs.existsArg("s", "search") {
		if cmdArgs.existsArg("q", "quiet") {
			config.searchMode = minimal
		} else {
			config.searchMode = detailed
		}
		return syncSearch(targets)
	}
	if cmdArgs.existsArg("p", "print", "print-format") {
		return show(passToPacman(cmdArgs))
	}
	if cmdArgs.existsArg("c", "clean") {
		return syncClean(cmdArgs)
	}
	if cmdArgs.existsArg("l", "list") {
		return show(passToPacman(cmdArgs))
	}
	if cmdArgs.existsArg("g", "groups") {
		return show(passToPacman(cmdArgs))
	}
	if cmdArgs.existsArg("i", "info") {
		return syncInfo(targets)
	}
	if cmdArgs.existsArg("u", "sysupgrade") {
		return install(cmdArgs)
	}
	if len(cmdArgs.targets) > 0 {
		return install(cmdArgs)
	}
	if cmdArgs.existsArg("y", "refresh") {
		return show(passToPacman(cmdArgs))
	}
	return nil
}

func handleRemove() error {
	removeVCSPackage(cmdArgs.targets)
	return show(passToPacman(cmdArgs))
}

// NumberMenu presents a CLI for selecting packages to install.
func displayNumberMenu(pkgS []string) (err error) {
	var (
		aurErr, repoErr error
		aq              aurQuery
		pq              repoQuery
		lenaq, lenpq    int
	)

	pkgS = removeInvalidTargets(pkgS)

	if config.mode == modeAUR || config.mode == modeAny {
		aq, aurErr = narrowSearch(pkgS, true)
		lenaq = len(aq)
	}
	if config.mode == modeRepo || config.mode == modeAny {
		pq, repoErr = queryRepo(pkgS)
		lenpq = len(pq)
		if repoErr != nil {
			return err
		}
	}

	if lenpq == 0 && lenaq == 0 {
		return fmt.Errorf("No packages match search")
	}

	if config.value["sortmode"] == "bottomup" {
		if config.mode == modeAUR || config.mode == modeAny {
			aq.printSearch(lenpq + 1)
		}
		if config.mode == modeRepo || config.mode == modeAny {
			pq.printSearch()
		}
	} else {
		if config.mode == modeRepo || config.mode == modeAny {
			pq.printSearch()
		}
		if config.mode == modeAUR || config.mode == modeAny {
			aq.printSearch(lenpq + 1)
		}
	}

	if aurErr != nil {
		fmt.Printf("Error during AUR search: %s\n", aurErr)
		fmt.Println("Showing repo packages only")
	}

	fmt.Println(bold(green(arrow + " Packages to install (eg: 1 2 3, 1-3 or ^4)")))
	fmt.Print(bold(green(arrow + " ")))

	reader := bufio.NewReader(os.Stdin)

	numberBuf, overflow, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if overflow {
		return fmt.Errorf("Input too long")
	}

	include, exclude, _, otherExclude := parseNumberMenu(string(numberBuf))
	arguments := makeArguments()

	isInclude := len(exclude) == 0 && len(otherExclude) == 0

	for i, pkg := range pq {
		target := len(pq) - i
		if config.value["sortmode"] == "topdown" {
			target = i + 1
		}

		if (isInclude && include.get(target)) || (!isInclude && !exclude.get(target)) {
			arguments.addTarget(pkg.DB().Name() + "/" + pkg.Name())
		}
	}

	for i, pkg := range aq {
		target := len(aq) - i + len(pq)
		if config.value["sortmode"] == "topdown" {
			target = i + 1 + len(pq)
		}

		if (isInclude && include.get(target)) || (!isInclude && !exclude.get(target)) {
			arguments.addTarget("aur/" + pkg.Name)
		}
	}

	if len(arguments.targets) == 0 {
		fmt.Println("There is nothing to do")
		return nil
	}

	if config.boolean["sudoloop"] {
		sudoLoopBackground()
	}

	err = install(arguments)

	return err
}
