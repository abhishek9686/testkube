## kubectl-testkube watch execution

Watch logs output from executor pod

### Synopsis

Gets test execution details, until it's in success/error state, blocks until gets complete state

```
kubectl-testkube watch execution <executionID> [flags]
```

### Options

```
  -h, --help   help for execution
```

### Options inherited from parent commands

```
      --analytics-enabled   enable analytics (default true)
  -c, --client string       client used for connecting to Testkube API one of proxy|direct (default "proxy")
  -s, --namespace string    Kubernetes namespace, default value read from config if set (default "testkube")
  -v, --verbose             show additional debug messages
```

### SEE ALSO

* [kubectl-testkube watch](kubectl-testkube_watch.md)	 - Watch tests or test suites
