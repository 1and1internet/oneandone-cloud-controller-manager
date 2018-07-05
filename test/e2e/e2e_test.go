package e2e

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	"gopkg.in/yaml.v2"
)

const (
	cloudPanelAPIKeyEnvVar = "CLOUD_PANEL_API_KEY"
	k8sVersionEnvVar       = "K8S_VERSION"
	terraformE2E           = "test/e2e/terraform"
	ansibleE2E             = "test/e2e/ansible"
	defaultKubeVersion     = "v1.10.5"
)

var kubeVersion, terraformDir, ansibleDir, kubeconfig string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	terraformDir = filepath.Join(dir, terraformE2E)
	ansibleDir = filepath.Join(dir, ansibleE2E)
	kubeconfig = filepath.Join(ansibleDir, "kubeconfig")
}

type terraformOutput struct {
	ClusterName struct {
		Value string
	} `json:"cluster_name"`
	Master struct {
		Value struct {
			Hostname string `json:"hostname"`
			IP       string `json:"ip"`
		} `json:"value"`
	} `json:"master"`
	SSHPrivateKey struct {
		Value string `json:"value"`
	} `json:"ssh_private_key"`
	Workers struct {
		Value struct {
			Hostnames []string `json:"hostnames"`
			IPs       []struct {
				IP string `json:"ip"`
			}
		}
	} `json:"workers"`
}

type ansibleHostVars struct {
	AnsibleHost string `yaml:"ansible_host"`
	PrivateIP   string `yaml:"private_ip"`
}

type ansibleHostsFile struct {
	All struct {
		Children struct {
			Masters struct {
				Hosts map[string]ansibleHostVars
			}
			Workers struct {
				Hosts map[string]ansibleHostVars
			}
		}
		Vars struct {
			APIToken string `yaml:"api_token"`
		}
	}
}

func getTerraformOutput() (terraformOutput, error) {
	var output terraformOutput
	cmd := exec.Command("terraform", "output", "-json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return output, err
	}
	if err := cmd.Start(); err != nil {
		return output, err
	}
	if err := json.NewDecoder(stdout).Decode(&output); err != nil {
		return output, err
	}
	if err := cmd.Wait(); err != nil {
		return output, err
	}
	return output, nil
}

func writeSSHPrivateKeyFile(t terraformOutput) error {
	idRSAFile := path.Join(ansibleDir, "id_rsa")
	return ioutil.WriteFile(idRSAFile, []byte(t.SSHPrivateKey.Value), 0600)
}

func writeAnsibleHostsFile(t terraformOutput) error {
	hostsFile := ansibleHostsFile{}
	var token = os.Getenv(cloudPanelAPIKeyEnvVar)

	children := &hostsFile.All.Children
	children.Masters.Hosts = make(map[string]ansibleHostVars)
	children.Workers.Hosts = make(map[string]ansibleHostVars)

	privateIP, err := getPrivateIP(token, t.Master.Value.Hostname)
	if err != nil {
		return err
	}
	children.Masters.Hosts[t.Master.Value.Hostname] = ansibleHostVars{t.Master.Value.IP, privateIP}

	for i := 0; i < len(t.Workers.Value.Hostnames); i++ {
		hostname := t.Workers.Value.Hostnames[i]
		ip := t.Workers.Value.IPs[i].IP
		privateIP, err := getPrivateIP(token, hostname)
		if err != nil {
			return err
		}
		children.Workers.Hosts[hostname] = ansibleHostVars{ip, privateIP}
	}

	hostsFile.All.Vars.APIToken = token
	bytes, err := yaml.Marshal(hostsFile)
	if err != nil {
		return err
	}

	hostsFilename := path.Join(ansibleDir, "hosts.yaml")
	return ioutil.WriteFile(hostsFilename, bytes, 0644)
}

func runCommandStreamingStdout(cmd *exec.Cmd) error {
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err = cmd.Start(); err != nil {
		return err
	}
	bytes, _ := ioutil.ReadAll(errReader)
	if err = cmd.Wait(); err != nil {
		return errors.New(string(bytes))
	}

	return nil
}

func buildCluster() error {
	var (
		cmdOut []byte
		err    error
	)

	token := os.Getenv(cloudPanelAPIKeyEnvVar)
	terraformTokenVar := fmt.Sprintf("provider_token=%s", token)

	log.Println("Preparing Terraform")

	// Terraform init
	if err = os.Chdir(terraformDir); err != nil {
		return err
	}

	if cmdOut, err = exec.Command("terraform", "init").CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, cmdOut)
		return err
	}

	terraformApplyRequired := false

	// Terraform plan
	// -detailed-exitcode means we get 0 if no changes, 1 for error and 2 for changes
	if cmdOut, err = exec.Command("terraform", "plan", "-detailed-exitcode", "-var", terraformTokenVar).CombinedOutput(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode := ws.ExitStatus()
			switch exitCode {
			case 1:
				fmt.Fprintln(os.Stderr, string(cmdOut))
				return err
			case 2:
				terraformApplyRequired = true
			}
		}

	}

	if terraformApplyRequired {
		log.Println("Running Terraform")

		// Terraform apply
		if err = runCommandStreamingStdout(exec.Command("terraform", "apply", "-var", terraformTokenVar, "-auto-approve")); err != nil {
			return err
		}
	} else {
		log.Println("Skipping Terraform apply: resources are already created")
	}

	// Get Terraform output
	tfOut, err := getTerraformOutput()
	if err != nil {
		return err
	}
	log.Printf("Cluster name: %s\n", tfOut.ClusterName.Value)

	log.Println("Preparing Ansible")

	// Prep Ansible - write SSH private key file
	err = writeSSHPrivateKeyFile(tfOut)
	if err != nil {
		return err
	}

	// Prep Ansible - write hosts file
	err = writeAnsibleHostsFile(tfOut)
	if err != nil {
		return err
	}

	if err = os.Chdir(ansibleDir); err != nil {
		return err
	}

	log.Println("Running Ansible")

	// Run ansible
	extraVars := fmt.Sprintf("k8s_version=%s", kubeVersion)
	if err = runCommandStreamingStdout(exec.Command("ansible-playbook", "-i", "hosts.yaml", "--extra-vars", extraVars, "create-cluster.yaml")); err != nil {
		return err
	}

	// Replace server in kubeconfig with public IP address of the master
	if cmdOut, err = exec.Command("kubectl", "--kubeconfig", kubeconfig, "config", "set-cluster", "kubernetes", fmt.Sprintf("--server=https://%s:6443", tfOut.Master.Value.IP)).Output(); err != nil {
		fmt.Fprintln(os.Stderr, cmdOut)
		return err
	}

	initK8SClient()

	return nil
}

func deleteCluster() error {
	log.Println("Deleting cluster")

	token := os.Getenv(cloudPanelAPIKeyEnvVar)
	terraformTokenVar := fmt.Sprintf("provider_token=%s", token)

	if err := os.Chdir(terraformDir); err != nil {
		return err
	}

	// Note: terraform > v0.11.0 is replacing -force with -auto-approve
	return runCommandStreamingStdout(exec.Command("terraform", "destroy", "-var", terraformTokenVar, "-force"))
}

func TestNodes(t *testing.T) {
	log.Println("Testing nodes")

	timeout := time.After(180 * time.Second)
	tick := time.Tick(5 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("Timed out: node(s) have cloud provider uninitialized taint or are not ready")
		case <-tick:
			nodes, err := getNodes()
			if err != nil {
				t.Fatal(err)
			}

			nodesUntainted := true
			nodesReady := true

			for _, node := range nodes {
				for _, taint := range node.Spec.Taints {
					if taint.Key == "node.cloudprovider.kubernetes.io/uninitialized" {
						nodesUntainted = false
						continue
					}
				}
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
						nodesReady = false
						continue
					}
				}
			}

			if nodesUntainted && nodesReady {
				return
			}
		}
	}
}

func TestLB(t *testing.T) {
	log.Println("Testing loadbalancer")

	if err := createNamespace("lbtest"); err != nil {
		t.Fatal(err)
	}
	defer deleteNamespace("lbtest")

	if err := createNginxDeployment(); err != nil {
		t.Fatal(err)
	}
	if err := createNginxService(); err != nil {
		t.Fatal(err)
	}
	ip, err := getSvcExternalIP("lbtest", "nginx", 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ip == "" {
		t.Fatal("Failed to get external IP address for load-balanced service")
	}

	// We'll wait up to 10s to get a 200 from the LB endpoint
	// before validating page content
	var resp *http.Response
	timeout := time.After(10 * time.Second)
	tick := time.Tick(1 * time.Second)
loop:
	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for 200 response from load balancer external IP")
		case <-tick:
			resp, err = http.Get("http://" + ip)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode == 200 {
				defer resp.Body.Close()
				break loop
			} else {
				resp.Body.Close()
			}
		}
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "Welcome to nginx!") {
		t.Fatal("Could not validate text on Nginx holding page")
	}
}

func TestMain(m *testing.M) {
	if os.Getenv(cloudPanelAPIKeyEnvVar) == "" {
		fmt.Fprintf(os.Stderr, "Cannot run tests: environment variable %s not set\n", cloudPanelAPIKeyEnvVar)
		os.Exit(1)
	}

	flag.StringVar(&kubeVersion, "kubever", defaultKubeVersion, "Kubernetes version, e.g. v1.10.5")
	flag.Parse()
	if kubeVersion == defaultKubeVersion && os.Getenv(k8sVersionEnvVar) != "" {
		kubeVersion = os.Getenv(k8sVersionEnvVar)
	}
	log.Printf("Using Kubernetes version %s\n", kubeVersion)

	exitCode := func() int {
		if os.Getenv("CCM_E2E_SKIP_CLUSTER_DELETE") != "true" {
			defer func() {
				err := deleteCluster()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}()
		}
		err := buildCluster()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return m.Run()
	}()

	os.Exit(exitCode)
}
