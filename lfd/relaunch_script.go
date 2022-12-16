package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runPythonScript(script string) (string, error) {
	// cmd := exec.Command("python", "-c", "import appscript")
	test_command := "go run /Users/vignesh/school/BRDS/18749-brds-project/server/main.go 1 10 0 " + "jjs-macbook-pro.wifi.local.cmu.edu"
	cmd := exec.Command("python3", "-c", "import appscript; appscript.app('Terminal').do_script('"+test_command+"')")
	// cmd2 := exec.Command("python", "-c", "appscript.app('Terminal').do_script('go run /Users/vignesh/school/BRDS/18749-brds-project/server/main.go 1 10 0 jjs-macbook-pro.wifi.local.cmu.edu')")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	// err2 := cmd2.Run()
	// if err2 != nil {
	// 	return "", err2
	// }

	return out.String(), nil
}

func main() {
	output, err := runPythonScript("import appscript")
	// output1, err1 := runPythonScript("appscript.app('Terminal').do_script('go run /Users/vignesh/school/BRDS/18749-brds-project/server/main.go 1 10 0 jjs-macbook-pro.wifi.local.cmu.edu')")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(output)
	}
	// output1, err1 := runPythonScript("appscript.app('Terminal').do_script('go run /Users/vignesh/school/BRDS/18749-brds-project/server/main.go 1 10 0 jjs-macbook-pro.wifi.local.cmu.edu')")
	// if err1 != nil {
	// 	fmt.Println(err1)
	// } else {
	// 	fmt.Println(output1)
	// }
}
