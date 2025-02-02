## kubectl-testkube create

Create resource

```
kubectl-testkube create <resourceName> [flags]
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
      --analytics-enabled   enable analytics (default true)
  -c, --client string       client used for connecting to Testkube API one of proxy|direct (default "proxy")
  -s, --namespace string    Kubernetes namespace, default value read from config if set (default "testkube")
  -v, --verbose             show additional debug messages
```

### SEE ALSO

* [kubectl-testkube](kubectl-testkube.md)	 - Testkube entrypoint for kubectl plugin
* [kubectl-testkube create executor](kubectl-testkube_create_executor.md)	 - Create new Executor
* [kubectl-testkube create test](kubectl-testkube_create_test.md)	 - Create new Test
* [kubectl-testkube create testsuite](kubectl-testkube_create_testsuite.md)	 - Create new TestSuite
* [kubectl-testkube create webhook](kubectl-testkube_create_webhook.md)	 - Create new Webhook

