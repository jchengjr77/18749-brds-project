import subprocess
import appscript

def run_go_program(program_name):
  # Run the Go program and get the output
#   subprocess.call(['open', '-W', '-a', 'Terminal.app', 'go', 'run', program_name, '1', '10', '0', 'jjs-macbook-pro.wifi.local.cmu.edu'])
#   subprocess.call(['open', '-W', '-a', 'Terminal.app', 'python', '--args', '/Users/vignesh/school/BRDS/18749-brds-project/lfd/hello.py'])
#   output = subprocess.run(["go", "run", program_name, "1", "10", "0",  "jjs-macbook-pro.wifi.local.cmu.edu"], capture_output=True).stdout
#   print(output)
  appscript.app('Terminal').do_script('go run /Users/vignesh/school/BRDS/18749-brds-project/server/main.go 1 10 0 jjs-macbook-pro.wifi.local.cmu.edu')
  
run_go_program("server/main.go")

# from subprocess import call
# call(["gnome-terminal", "-x", "sh", "-c", "python3; bash"])
