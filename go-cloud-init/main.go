package main

import (
	"fmt"
	"github.com/juju/juju/cloudconfig/cloudinit"
	"log"
	"os"
)


func main() {
	cfg, err := cloudinit.New("xenial")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	cfg.SetAttr("timezone", "Asia/Tokyo")
	cfg.AddUser(&cloudinit.User{Name: "r_takaishi"})
	cfg.AddPackage("vim")

	cfg.AddRunTextFile("run_text_file_1", "aaa", 1)
	cfg.AddRunTextFile("run_text_file_2", "bbb", 2)

	cfg.AddRunCmd("add_run_cmd_1", "args")
	cfg.AddRunCmd("add_run_cmd_2", "args")

	cfg.AddBootTextFile("boot_text_file_1", "ccc", 3)
	cfg.AddBootTextFile("boot_text_file_2", "ddd", 4)

	cfg.AddBootCmd("add_boot_cmd_1", "args", )
	cfg.AddBootCmd("add_boot_cmd_2", "args", )

	cfg.AddScripts("add_scriptsccc")
	b, err := cfg.RenderYAML()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}
