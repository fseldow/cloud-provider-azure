export KUBECONFIG=/home/t-xinhli/new-provider/cloud-provider-azure/t-xinhli-new3/kubeconfig
alias km='ssh k8s-ci@t-xinhli-new3.eastus.cloudapp.azure.com'

# export KUBERNETES_PROVIDER=azure # no use when conformance=y
export KUBERNETES_CONFORMANCE_TEST=y
export KUBERNETES_CONFORMANCE_PROVIDER=azure
export CLOUD_CONFIG=1 # workaround for the new parameter.
