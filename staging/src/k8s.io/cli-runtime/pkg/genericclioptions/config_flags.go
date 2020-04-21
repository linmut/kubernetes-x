/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package genericclioptions

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	flagClusterName      = "cluster"
	flagAuthInfoName     = "user"
	flagContext          = "context"
	flagNamespace        = "namespace"
	flagAPIServer        = "server"
	flagInsecure         = "insecure-skip-tls-verify"
	flagCertFile         = "client-certificate"
	flagKeyFile          = "client-key"
	flagCAFile           = "certificate-authority"
	flagBearerToken      = "token"
	flagImpersonate      = "as"
	flagImpersonateGroup = "as-group"
	flagUsername         = "username"
	flagPassword         = "password"
	flagTimeout          = "request-timeout"
	flagHTTPCacheDir     = "cache-dir"
	flagOpensslKey       = "openssl-key"
)

var defaultCacheDir = filepath.Join(homedir.HomeDir(), ".kube", "http-cache")

// RESTClientGetter is an interface that the ConfigFlags describe to provide an easier way to mock for commands
// and eliminate the direct coupling to a struct type.  Users may wish to duplicate this type in their own packages
// as per the golang type overlapping.
type RESTClientGetter interface {
	// ToRESTConfig returns restconfig
	ToRESTConfig() (*rest.Config, error)
	// ToDiscoveryClient returns discovery client
	ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error)
	// ToRESTMapper returns a restmapper
	ToRESTMapper() (meta.RESTMapper, error)
	// ToRawKubeConfigLoader return kubeconfig loader as-is
	ToRawKubeConfigLoader() clientcmd.ClientConfig
}

var _ RESTClientGetter = &ConfigFlags{}

// ConfigFlags composes the set of values necessary
// for obtaining a REST client config
type ConfigFlags struct {
	CacheDir   *string
	KubeConfig *string

	//
	OpensslKey *string
	// config flags
	ClusterName      *string
	AuthInfoName     *string
	Context          *string
	Namespace        *string
	APIServer        *string
	Insecure         *bool
	CertFile         *string
	KeyFile          *string
	CAFile           *string
	BearerToken      *string
	Impersonate      *string
	ImpersonateGroup *[]string
	Username         *string
	Password         *string
	Timeout          *string

	clientConfig clientcmd.ClientConfig
	lock         sync.Mutex
	// If set to true, will use persistent client config and
	// propagate the config to the places that need it, rather than
	// loading the config multiple times
	usePersistentConfig bool
}

// ToRESTConfig implements RESTClientGetter.
// Returns a REST client configuration based on a provided path
// to a .kubeconfig file, loading rules, and config flag overrides.
// Expects the AddFlags method to have been called.
func (f *ConfigFlags) ToRESTConfig() (*rest.Config, error) {
	//TODO: todo for mulin, 在这里添加一个方法，负责解析参数和生产新的kubeconfig路径，It would be better
	return f.ToRawKubeConfigLoader().ClientConfig()
}

// ToRawKubeConfigLoader binds config flag values to config overrides
// Returns an interactive clientConfig if the password flag is enabled,
// or a non-interactive clientConfig otherwise.
func (f *ConfigFlags) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	//TODO: todo for mulin, 什么时候会触发这个分支？
	if f.usePersistentConfig {
		return f.toRawKubePersistentConfigLoader()
	}
	return f.toRawKubeConfigLoader()
}

//TODO: modified by mulin, 解析参数，并生产新的kubeconfig路径
func (f *ConfigFlags) toRawKubeConfigLoader() clientcmd.ClientConfig {
	//TODO: begin，调用栈
	//label := "toRawKubeConfigLoader"
	//file := fmt.Sprintf("/data/home/mulin/k8s-debug/%s_%s.txt", label, time.Now().Format("2006-01-02_15:04:05.000"))
	//logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	//if nil != err {
	//	panic(err)
	//}
	//loger := log.New(logFile, "前缀", log.Ldate|log.Ltime|log.Lshortfile)
	//loger.Printf("%s 调用栈 :%s", label, debug.Stack())
	//fmt.Printf("### 调用栈 %s\n", label)
	//TODO：end

	//TODO: begin, added by mulin, for opensslKey
	if f.OpensslKey != nil {
		fmt.Printf("### CALL toRawKubeConfigLoader, f.OpensslKey != nil\n")
	} else{
		fmt.Printf("### CALL toRawKubeConfigLoader, f.OpensslKey == nil\n")
	}
	fmt.Printf("### CALL toRawKubeConfigLoader, *f.OpensslKey=%s\n", *f.OpensslKey)
	//TODO: end


	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	// 移动到判断完namespace和cluster之后
	//if f.KubeConfig != nil {
	//	loadingRules.ExplicitPath = *f.KubeConfig
	//}

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	// bind auth info flag values to overrides
	if f.CertFile != nil {
		overrides.AuthInfo.ClientCertificate = *f.CertFile
	}
	if f.KeyFile != nil {
		overrides.AuthInfo.ClientKey = *f.KeyFile
	}
	if f.BearerToken != nil {
		overrides.AuthInfo.Token = *f.BearerToken
	}
	if f.Impersonate != nil {
		overrides.AuthInfo.Impersonate = *f.Impersonate
	}
	if f.ImpersonateGroup != nil {
		overrides.AuthInfo.ImpersonateGroups = *f.ImpersonateGroup
	}
	if f.Username != nil {
		overrides.AuthInfo.Username = *f.Username
	}
	if f.Password != nil {
		overrides.AuthInfo.Password = *f.Password
	}

	// bind cluster flags
	if f.APIServer != nil {
		overrides.ClusterInfo.Server = *f.APIServer
	}
	if f.CAFile != nil {
		overrides.ClusterInfo.CertificateAuthority = *f.CAFile
	}
	if f.Insecure != nil {
		overrides.ClusterInfo.InsecureSkipTLSVerify = *f.Insecure
	}

	// bind context flags
	if f.Context != nil {
		overrides.CurrentContext = *f.Context
	}

	//TODO: begin, modified by mulin, 额外判断
	//if f.ClusterName != nil {
	//	overrides.Context.Cluster = *f.ClusterName
	//}
	if len(*f.ClusterName) == 0 {
		err := fmt.Errorf("--cluster= 参数不能为空，必须指定集群id，前缀为'cls-'")
		panic(err)
	} else {
		fmt.Printf("### CALL toRawKubeConfigLoader, *f.ClusterName=%s\n", *f.ClusterName)
		overrides.Context.Cluster = *f.ClusterName
	}
	//TODO: end

	if f.AuthInfoName != nil {
		overrides.Context.AuthInfo = *f.AuthInfoName
	}

	//TODO: begin, modified by mulin, 额外判断
	//if f.Namespace != nil {
	//	overrides.Context.Namespace = *f.Namespace
	//}
	if len(*f.Namespace) == 0 {
		err := fmt.Errorf("-n 或--namespace= 参数不能为空，必须指定业务的命名空间，前缀为'ns-'")
		panic(err)
	} else {
		fmt.Printf("### CALL toRawKubeConfigLoader, *f.Namespace=%s\n", *f.Namespace)
		overrides.Context.Namespace = *f.Namespace
	}
	//TODO: end

	if f.Timeout != nil {
		overrides.Timeout = *f.Timeout
	}

	//TODO: begin, modified by mulin, 额外判断
	//if f.KubeConfig != nil {
	//	loadingRules.ExplicitPath = *f.KubeConfig
	//}
	if f.KubeConfig != nil {
		if len(*f.KubeConfig) > 0 {
			err := fmt.Errorf("--kubeconfig= 参数不能非空(%s)，不支持指定特定的kubeconfig", *f.KubeConfig)
			panic(err)
		}
	}
	cluster := *f.ClusterName
	namespace := *f.Namespace
	oaUser, err := user.Current()
	if err != nil {
		err = fmt.Errorf("get current user failed, err=%s\n", err.Error())
		panic(err)
	} else {
		if oaUser.Username == "root" {
			err = fmt.Errorf("当前用户username=%s执行异常，请联系TkeHelper\n", oaUser.Username)
			panic(err)
		}
	}
	myKubeconfig := oaUser.HomeDir + "/.kube/" + cluster + "_" + namespace + "_" + oaUser.Username + ".kubeconfig"
	fmt.Printf("### CALL myKubeconfig=%s\n", myKubeconfig)
	loadingRules.ExplicitPath = myKubeconfig
	//TODO: end

	var clientConfig clientcmd.ClientConfig
	// we only have an interactive prompt when a password is allowed
	if f.Password == nil {
		clientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	} else {
		clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)
	}

	return clientConfig
}

// toRawKubePersistentConfigLoader binds config flag values to config overrides
// Returns a persistent clientConfig for propagation.
func (f *ConfigFlags) toRawKubePersistentConfigLoader() clientcmd.ClientConfig {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.clientConfig == nil {
		f.clientConfig = f.toRawKubeConfigLoader()
	}

	return f.clientConfig
}

// ToDiscoveryClient implements RESTClientGetter.
// Expects the AddFlags method to have been called.
// Returns a CachedDiscoveryInterface using a computed RESTConfig.
func (f *ConfigFlags) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom resources) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = 100

	// retrieve a user-provided value for the "cache-dir"
	// defaulting to ~/.kube/http-cache if no user-value is given.
	httpCacheDir := defaultCacheDir
	if f.CacheDir != nil {
		httpCacheDir = *f.CacheDir
	}

	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), config.Host)
	return diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, time.Duration(10*time.Minute))
}

// ToRESTMapper returns a mapper.
func (f *ConfigFlags) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// AddFlags binds client configuration flags to a given flagset
func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet) {
	//TODO: begin, added by mulin, for opensslKey
	if f.OpensslKey != nil {
		//fmt.Printf("### CALL AddFlags, 1, *f.KubeConfig=%s\n", *f.KubeConfig)
		flags.StringVar(f.OpensslKey, flagOpensslKey, *f.OpensslKey, "openssl key.")
	}
	//TODO: end

	if f.KubeConfig != nil {
		//fmt.Printf("### CALL AddFlags, 1, *f.KubeConfig=%s\n", *f.KubeConfig)
		flags.StringVar(f.KubeConfig, "kubeconfig", *f.KubeConfig, "Path to the kubeconfig file to use for CLI requests.")
	}
	if f.CacheDir != nil {
		//fmt.Printf("### CALL AddFlags, 2, *f.CacheDir=%s\n", *f.CacheDir)
		flags.StringVar(f.CacheDir, flagHTTPCacheDir, *f.CacheDir, "Default HTTP cache directory")
	}

	// add config options
	if f.CertFile != nil {
		//fmt.Printf("### CALL AddFlags, 3, *f.CertFile=%s\n", *f.CertFile)
		flags.StringVar(f.CertFile, flagCertFile, *f.CertFile, "Path to a client certificate file for TLS")
	}
	if f.KeyFile != nil {
		//fmt.Printf("### CALL AddFlags, 4, *f.KeyFile=%s\n", *f.KeyFile)
		flags.StringVar(f.KeyFile, flagKeyFile, *f.KeyFile, "Path to a client key file for TLS")
	}
	if f.BearerToken != nil {
		//fmt.Printf("### CALL AddFlags, 5, *f.BearerToken=%s\n", *f.BearerToken)
		flags.StringVar(f.BearerToken, flagBearerToken, *f.BearerToken, "Bearer token for authentication to the API server")
	}
	if f.Impersonate != nil {
		//fmt.Printf("### CALL AddFlags, 6, *f.Impersonate=%s\n", *f.Impersonate)
		flags.StringVar(f.Impersonate, flagImpersonate, *f.Impersonate, "Username to impersonate for the operation")
	}
	if f.ImpersonateGroup != nil {
		//fmt.Printf("### CALL AddFlags, 7, *f.ImpersonateGroup=%s\n", *f.ImpersonateGroup)
		flags.StringArrayVar(f.ImpersonateGroup, flagImpersonateGroup, *f.ImpersonateGroup, "Group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	}
	if f.Username != nil {
		//fmt.Printf("### CALL AddFlags, 8, *f.Username=%s\n", *f.Username)
		flags.StringVar(f.Username, flagUsername, *f.Username, "Username for basic authentication to the API server")
	}
	if f.Password != nil {
		//fmt.Printf("### CALL AddFlags, 9, *f.Password=%s\n", *f.Password)
		flags.StringVar(f.Password, flagPassword, *f.Password, "Password for basic authentication to the API server")
	}
	if f.ClusterName != nil {
		//fmt.Printf("### CALL AddFlags, 10, *f.ClusterName=%s\n", *f.ClusterName)
		flags.StringVar(f.ClusterName, flagClusterName, *f.ClusterName, "The name of the kubeconfig cluster to use")
	}
	if f.AuthInfoName != nil {
		//fmt.Printf("### CALL AddFlags, 11, *f.AuthInfoName=%s\n", *f.AuthInfoName)
		flags.StringVar(f.AuthInfoName, flagAuthInfoName, *f.AuthInfoName, "The name of the kubeconfig user to use")
	}
	if f.Namespace != nil {
		//fmt.Printf("### CALL AddFlags, 12, *f.Namespace=%s\n", *f.Namespace)
		flags.StringVarP(f.Namespace, flagNamespace, "n", *f.Namespace, "If present, the namespace scope for this CLI request")
	}
	if f.Context != nil {
		//fmt.Printf("### CALL AddFlags, 13, *f.Context=%s\n", *f.Context)
		flags.StringVar(f.Context, flagContext, *f.Context, "The name of the kubeconfig context to use")
	}

	if f.APIServer != nil {
		//fmt.Printf("### CALL AddFlags, 14, *f.APIServer=%s\n", *f.APIServer)
		flags.StringVarP(f.APIServer, flagAPIServer, "s", *f.APIServer, "The address and port of the Kubernetes API server")
	}
	if f.Insecure != nil {
		//fmt.Printf("### CALL AddFlags, 15, *f.Insecure=%s\n", *f.Insecure)
		flags.BoolVar(f.Insecure, flagInsecure, *f.Insecure, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	}
	if f.CAFile != nil {
		//fmt.Printf("### CALL AddFlags, 16, *f.CAFile=%s\n", *f.CAFile)
		flags.StringVar(f.CAFile, flagCAFile, *f.CAFile, "Path to a cert file for the certificate authority")
	}
	if f.Timeout != nil {
		//fmt.Printf("### CALL AddFlags, 17, *f.Timeout=%s\n", *f.Timeout)
		flags.StringVar(f.Timeout, flagTimeout, *f.Timeout, "The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests.")
	}

}

// WithDeprecatedPasswordFlag enables the username and password config flags
func (f *ConfigFlags) WithDeprecatedPasswordFlag() *ConfigFlags {
	f.Username = stringptr("")
	f.Password = stringptr("")
	return f
}

// NewConfigFlags returns ConfigFlags with default values set
func NewConfigFlags(usePersistentConfig bool) *ConfigFlags {
	impersonateGroup := []string{}
	insecure := false

	return &ConfigFlags{
		Insecure:   &insecure,
		Timeout:    stringptr("0"),
		KubeConfig: stringptr(""),

		CacheDir:         stringptr(defaultCacheDir),
		ClusterName:      stringptr(""),
		AuthInfoName:     stringptr(""),
		Context:          stringptr(""),
		Namespace:        stringptr(""),
		APIServer:        stringptr(""),
		CertFile:         stringptr(""),
		KeyFile:          stringptr(""),
		CAFile:           stringptr(""),
		BearerToken:      stringptr(""),
		Impersonate:      stringptr(""),
		ImpersonateGroup: &impersonateGroup,

		usePersistentConfig: usePersistentConfig,
	}
}

func stringptr(val string) *string {
	return &val
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}
