# SoapUI Tests

[SoapUI](https://www.soapui.org) is an open-source tool used for end-to-end testing of REST, SOAP, & GraphQL APIs, JMS, JDBC, and other web services. Testkube supports the usage of it with the SoapUI executor implementation.

## Running a SoapUI test

In order to run a SoapUI test using Testkube, it is necessary to create a Testkube Test.

### Using files as input

Testkube and the SoapUI executor accepts a project file as input.

```sh
$ kubectl testkube create test --file REST-Project-1-soapui-project.xml --type soapui/rest --name example-test

████████ ███████ ███████ ████████ ██   ██ ██    ██ ██████  ███████ 
   ██    ██      ██         ██    ██  ██  ██    ██ ██   ██ ██      
   ██    █████   ███████    ██    █████   ██    ██ ██████  █████   
   ██    ██           ██    ██    ██  ██  ██    ██ ██   ██ ██      
   ██    ███████ ███████    ██    ██   ██  ██████  ██████  ███████ 
                                           /tɛst kjub/ by Kubeshop


Test created  / example-test 🥇

```

### Using strings as input

```sh
$ cat REST-Project-1-soapui-project.xml | kubectl testkube create test --type soapui/rest --name example-test-string

████████ ███████ ███████ ████████ ██   ██ ██    ██ ██████  ███████ 
   ██    ██      ██         ██    ██  ██  ██    ██ ██   ██ ██      
   ██    █████   ███████    ██    █████   ██    ██ ██████  █████   
   ██    ██           ██    ██    ██  ██  ██    ██ ██   ██ ██      
   ██    ███████ ███████    ██    ██   ██  ██████  ██████  ███████ 
                                           /tɛst kjub/ by Kubeshop


Test created  / example-test-string 🥇

```

### Running the tests

To run the created test, use:

```sh
$ kubectl testkube run test example-test

████████ ███████ ███████ ████████ ██   ██ ██    ██ ██████  ███████ 
   ██    ██      ██         ██    ██  ██  ██    ██ ██   ██ ██      
   ██    █████   ███████    ██    █████   ██    ██ ██████  █████   
   ██    ██           ██    ██    ██  ██  ██    ██ ██   ██ ██      
   ██    ███████ ███████    ██    ██   ██  ██████  ██████  ███████ 
                                           /tɛst kjub/ by Kubeshop


Type          : soapui/rest
Name          : example-test
Execution ID  : 624eedd443ed8485ae9289e2
Execution name: illegally-credible-mouse



Test execution started

Watch test execution until complete:
$ kubectl testkube watch execution 624eedd443ed8485ae9289e2


Use following command to get test execution details:
$ kubectl testkube get execution 624eedd443ed8485ae9289e2

```

### Using parameters and arguments in your tests

SoapUI lets you configure your test runs using different parameters. To see all available command line arguments, check the [official SoapUI docs](https://www.soapui.org/docs/test-automation/running-functional-tests/).

When working with Testkube, the way to use the parameters is by using the `kubectl testkube start` command with the `--args` parameter.
An example would be:

```sh
$ kubectl testkube start test -f example-test --args '-I -c "TestCase 1"'

████████ ███████ ███████ ████████ ██   ██ ██    ██ ██████  ███████ 
   ██    ██      ██         ██    ██  ██  ██    ██ ██   ██ ██      
   ██    █████   ███████    ██    █████   ██    ██ ██████  █████   
   ██    ██           ██    ██    ██  ██  ██    ██ ██   ██ ██      
   ██    ███████ ███████    ██    ██   ██  ██████  ██████  ███████ 
                                           /tɛst kjub/ by Kubeshop


Type          : soapui/rest
Name          : successful-test
Execution ID  : 625404e5a4cc6d2861193c60
Execution name: currently-amused-pug


Getting pod logs
Execution completed ================================
=
= SOAPUI_HOME = /usr/local/SmartBear/SoapUI-5.7.0
=
================================
SoapUI 5.7.0 TestCase Runner
10:37:37,713 INFO  [DefaultSoapUICore] Creating new settings at [/root/soapui-settings.xml]
10:37:43,567 INFO  [PluginManager] 0 plugins loaded in 36 ms
10:37:43,570 INFO  [DefaultSoapUICore] All plugins loaded
10:37:50,774 INFO  [WsdlProject] Loaded project from [file:/tmp/test-content359342991]
10:37:50,834 INFO  [SoapUITestCaseRunner] Running SoapUI tests in project [REST Project 2]
10:37:50,838 INFO  [SoapUITestCaseRunner] Running TestCase [TestCase 1]
10:37:50,876 INFO  [SoapUITestCaseRunner] Running SoapUI testcase [TestCase 1]
10:37:50,901 INFO  [SoapUITestCaseRunner] running step [1 - Request 1]
10:37:54,180 INFO  [SoapUITestCaseRunner] Assertion [Valid HTTP Status Codes] has status VALID
10:37:54,193 INFO  [SoapUITestCaseRunner] Assertion [Contains] has status VALID
10:37:54,257 INFO  [SoapUITestCaseRunner] Finished running SoapUI testcase [TestCase 1], time taken: 990ms, status: FINISHED
10:37:54,315 INFO  [SoapUITestCaseRunner] TestCase [TestCase 1] finished with status [FINISHED] in 990ms


.
Use following command to get test execution details:
$ kubectl testkube get execution 625404e5a4cc6d2861193c60
```

Usage of the `-I` argument is highly suggested to get cleaner results.

## Reports, plugins and extensions

In order to be able to use reports, add plugins and extensions the way [SoapUI docs](https://www.soapui.org/docs/test-automation/running-in-docker/) describe it, is currently not supported by Testkube.
In case you need this feature, please create an [issue](https://github.com/kubeshop/testkube/issues) in the [Testkube repository](https://github.com/kubeshop/testkube).
