package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testsuitesv1 "github.com/kubeshop/testkube-operator/apis/testsuite/v1"
	"github.com/kubeshop/testkube/internal/pkg/api/datefilter"
	"github.com/kubeshop/testkube/internal/pkg/api/repository/testresult"
	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/cronjob"
	testsuitesmapper "github.com/kubeshop/testkube/pkg/mapper/testsuites"
	"github.com/kubeshop/testkube/pkg/rand"
	"github.com/kubeshop/testkube/pkg/types"
)

// GetTestSuiteHandler for getting test object
func (s TestkubeAPI) CreateTestSuiteHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request testkube.TestSuiteUpsertRequest
		err := c.BodyParser(&request)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		testSuite := mapTestSuiteUpsertRequestToTestCRD(request)
		testSuite.Namespace = s.Namespace

		created, err := s.TestsSuitesClient.Create(&testSuite)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		c.Status(201)
		return c.JSON(created)
	}
}

// UpdateTestSuiteHandler updates an existing TestSuite CR based on TestSuite content
func (s TestkubeAPI) UpdateTestSuiteHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request testkube.TestSuiteUpsertRequest
		err := c.BodyParser(&request)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		// we need to get resource first and load its metadata.ResourceVersion
		testSuite, err := s.TestsSuitesClient.Get(request.Name)
		if err != nil {
			return s.Error(c, http.StatusBadGateway, err)
		}

		// delete cron job, if schedule is cleaned
		if testSuite.Spec.Schedule != "" && request.Schedule == "" {
			if err = s.CronJobClient.Delete(cronjob.GetMetadataName(request.Name, testSuiteResourceURI)); err != nil {
				if !errors.IsNotFound(err) {
					return s.Error(c, http.StatusBadGateway, err)
				}
			}
		}

		// map TestSuite but load spec only to not override metadata.ResourceVersion
		testSuiteSpec := mapTestSuiteUpsertRequestToTestCRD(request)
		testSuite.Spec = testSuiteSpec.Spec
		testSuite.Labels = request.Labels
		testSuite, err = s.TestsSuitesClient.Update(testSuite)
		if err != nil {
			return s.Error(c, http.StatusBadGateway, err)
		}

		return c.JSON(testSuite)
	}
}

// GetTestSuiteHandler for getting TestSuite object
func (s TestkubeAPI) GetTestSuiteHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("id")
		crTestSuite, err := s.TestsSuitesClient.Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				return s.Warn(c, http.StatusNotFound, err)
			}

			return s.Error(c, http.StatusBadGateway, err)
		}

		testSuite := testsuitesmapper.MapCRToAPI(*crTestSuite)

		return c.JSON(testSuite)
	}
}

// GetTestSuiteWithExecutionHandler for getting TestSuite object with execution
func (s TestkubeAPI) GetTestSuiteWithExecutionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("id")
		crTestSuite, err := s.TestsSuitesClient.Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				return s.Warn(c, http.StatusNotFound, err)
			}

			return s.Error(c, http.StatusBadGateway, err)
		}

		ctx := c.Context()
		execution, err := s.TestExecutionResults.GetLatestByTest(ctx, name)
		if err != nil && err != mongo.ErrNoDocuments {
			return s.Error(c, http.StatusInternalServerError, err)
		}

		testSuite := testsuitesmapper.MapCRToAPI(*crTestSuite)
		testSuiteWithExecution := testkube.TestSuiteWithExecution{
			TestSuite: &testSuite,
		}
		if err == nil {
			testSuiteWithExecution.LatestExecution = &execution
		}

		return c.JSON(testSuiteWithExecution)
	}
}

// DeleteTestSuiteHandler for deleting a TestSuite with id
func (s TestkubeAPI) DeleteTestSuiteHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("id")
		err := s.TestsSuitesClient.Delete(name)
		if err != nil {
			if errors.IsNotFound(err) {
				return s.Warn(c, http.StatusNotFound, err)
			}

			return s.Error(c, http.StatusBadGateway, err)
		}

		// delete cron job for test suite
		if err = s.CronJobClient.Delete(cronjob.GetMetadataName(name, testSuiteResourceURI)); err != nil {
			if !errors.IsNotFound(err) {
				return s.Error(c, http.StatusBadGateway, err)
			}
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// DeleteTestSuitesHandler for deleting all TestSuites
func (s TestkubeAPI) DeleteTestSuitesHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := s.TestsSuitesClient.DeleteAll()
		if err != nil {
			if errors.IsNotFound(err) {
				return s.Warn(c, http.StatusNotFound, err)
			}

			return s.Error(c, http.StatusBadGateway, err)
		}

		// delete all cron jobs for test suites
		if err = s.CronJobClient.DeleteAll(testSuiteResourceURI); err != nil {
			if !errors.IsNotFound(err) {
				return s.Error(c, http.StatusBadGateway, err)
			}
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func (s TestkubeAPI) getFilteredTestSuitesList(c *fiber.Ctx) (*testsuitesv1.TestSuiteList, error) {
	crTestSuites, err := s.TestsSuitesClient.List(c.Query("selector"))
	if err != nil {
		return nil, err
	}

	search := c.Query("textSearch")
	if search != "" {
		// filter items array
		for i := len(crTestSuites.Items) - 1; i >= 0; i-- {
			if !strings.Contains(crTestSuites.Items[i].Name, search) {
				crTestSuites.Items = append(crTestSuites.Items[:i], crTestSuites.Items[i+1:]...)
			}
		}
	}

	return crTestSuites, nil
}

// ListTestSuitesHandler for getting list of all available TestSuites
func (s TestkubeAPI) ListTestSuitesHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		crTestSuites, err := s.getFilteredTestSuitesList(c)
		if err != nil {
			return s.Error(c, http.StatusInternalServerError, err)
		}

		testSuites := testsuitesmapper.MapTestSuiteListKubeToAPI(*crTestSuites)

		return c.JSON(testSuites)
	}
}

// ListTestSuiteWithExecutionsHandler for getting list of all available TestSuite with latest executions
func (s TestkubeAPI) ListTestSuiteWithExecutionsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		crTestSuites, err := s.getFilteredTestSuitesList(c)
		if err != nil {
			return s.Error(c, http.StatusInternalServerError, err)
		}

		ctx := c.Context()
		testSuites := testsuitesmapper.MapTestSuiteListKubeToAPI(*crTestSuites)
		testSuiteWithExecutions := make([]testkube.TestSuiteWithExecution, len(testSuites))
		testNames := make([]string, len(testSuites))
		for i := range testSuites {
			testNames[i] = testSuites[i].Name
		}

		executions, err := s.TestExecutionResults.GetLatestByTests(ctx, testNames)
		if err != nil && err != mongo.ErrNoDocuments {
			return s.Error(c, http.StatusInternalServerError, err)
		}

		executionMap := make(map[string]testkube.TestSuiteExecution, len(executions))
		for i := range executions {
			if executions[i].TestSuite == nil {
				continue
			}

			executionMap[executions[i].TestSuite.Name] = executions[i]
		}

		for i := range testSuites {
			testSuiteWithExecutions[i].TestSuite = &testSuites[i]
			if execution, ok := executionMap[testSuites[i].Name]; ok {
				testSuiteWithExecutions[i].LatestExecution = &execution
			}
		}

		status := c.Query("status")
		if status != "" {
			// filter items array
			for i := len(testSuiteWithExecutions) - 1; i >= 0; i-- {
				if testSuiteWithExecutions[i].LatestExecution != nil && testSuiteWithExecutions[i].LatestExecution.Status != nil &&
					*testSuiteWithExecutions[i].LatestExecution.Status == testkube.TestSuiteExecutionStatus(status) {
					continue
				}

				testSuiteWithExecutions = append(testSuiteWithExecutions[:i], testSuiteWithExecutions[i+1:]...)
			}
		}

		return c.JSON(testSuiteWithExecutions)
	}
}

func (s TestkubeAPI) ExecuteTestSuiteHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		name := c.Params("id")
		namespace := c.Query("namespace", "testkube")
		s.Log.Debugw("getting test suite", "name", name)

		crTestSuite, err := s.TestsSuitesClient.Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				return s.Warn(c, http.StatusNotFound, err)
			}

			return s.Error(c, http.StatusBadGateway, err)
		}

		var request testkube.TestSuiteExecutionRequest
		err = c.BodyParser(&request)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, fmt.Errorf("test execution request body invalid: %w", err))
		}

		if crTestSuite.Spec.Schedule != "" && c.Query("callback") == "" {
			data, err := json.Marshal(request)
			if err != nil {
				return s.Error(c, http.StatusBadRequest, fmt.Errorf("can't prepare test suite request: %w", err))
			}

			options := cronjob.CronJobOptions{
				Schedule: crTestSuite.Spec.Schedule,
				Resource: testSuiteResourceURI,
				Data:     string(data),
			}
			if err = s.CronJobClient.Apply(crTestSuite.Name, cronjob.GetMetadataName(crTestSuite.Name, testSuiteResourceURI), options); err != nil {
				return s.Error(c, http.StatusInternalServerError, fmt.Errorf("can't create scheduled test suite: %w", err))
			}

			return c.JSON(testkube.TestSuiteExecution{
				TestSuite: &testkube.ObjectRef{
					Name:      name,
					Namespace: namespace,
				},
				Status: testkube.TestSuiteExecutionStatusQueued,
			})
		}

		testSuite := testsuitesmapper.MapCRToAPI(*crTestSuite)
		s.Log.Debugw("executing test", "name", name, "test suite", testSuite, "cr", crTestSuite)
		results := s.executeTestSuite(ctx, request, testSuite)

		c.Response().SetStatusCode(fiber.StatusCreated)
		return c.JSON(results)
	}
}

func (s TestkubeAPI) ListTestSuiteExecutionsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		filter := getExecutionsFilterFromRequest(c)

		executionsTotals, err := s.TestExecutionResults.GetExecutionsTotals(ctx, filter)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}
		allExecutionsTotals, err := s.TestExecutionResults.GetExecutionsTotals(ctx)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		executions, err := s.TestExecutionResults.GetExecutions(ctx, filter)
		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		return c.JSON(testkube.TestSuiteExecutionsResult{
			Totals:   &allExecutionsTotals,
			Filtered: &executionsTotals,
			Results:  mapToTestExecutionSummary(executions),
		})
	}
}

func (s TestkubeAPI) GetTestSuiteExecutionHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		id := c.Params("executionID")
		execution, err := s.TestExecutionResults.Get(ctx, id)

		if err != nil {
			return s.Error(c, http.StatusBadRequest, err)
		}

		execution.Duration = types.FormatDuration(execution.Duration)

		return c.JSON(execution)
	}
}

func (s TestkubeAPI) executeTestSuite(ctx context.Context, request testkube.TestSuiteExecutionRequest, test testkube.TestSuite) (testExecution testkube.TestSuiteExecution) {
	s.Log.Debugw("Got test to execute", "test", test)

	testExecution = testkube.NewStartedTestSuiteExecution(test, request)
	s.TestExecutionResults.Insert(ctx, testExecution)

	go func(testExecution testkube.TestSuiteExecution) {

		defer func(testExecution *testkube.TestSuiteExecution) {
			duration := testExecution.CalculateDuration()
			testExecution.EndTime = time.Now()
			testExecution.Duration = duration.String()

			err := s.TestExecutionResults.EndExecution(ctx, testExecution.Id, testExecution.EndTime, duration)
			if err != nil {
				s.Log.Errorw("error setting end time", "error", err.Error())
			}
		}(&testExecution)

		hasFailedSteps := false
		for i := range testExecution.StepResults {

			// start execution of given step
			testExecution.StepResults[i].Execution.ExecutionResult.InProgress()
			s.TestExecutionResults.Update(ctx, testExecution)

			s.executeTestStep(ctx, testExecution, &testExecution.StepResults[i])
			err := s.TestExecutionResults.Update(ctx, testExecution)
			if err != nil {
				hasFailedSteps = true
				s.Log.Errorw("saving test suite execution results error", "error", err)
				continue
			}

			if testExecution.StepResults[i].IsFailed() {
				hasFailedSteps = true
				if testExecution.StepResults[i].Step.StopTestOnFailure {
					break
				}
			}
		}

		testExecution.Status = testkube.TestSuiteExecutionStatusPassed
		if hasFailedSteps {
			testExecution.Status = testkube.TestSuiteExecutionStatusFailed
		}

		err := s.TestExecutionResults.Update(ctx, testExecution)
		if err != nil {
			s.Log.Errorw("saving final test suite execution result error", "error", err)
		}

	}(testExecution)

	return

}

func (s TestkubeAPI) executeTestStep(ctx context.Context, testExecution testkube.TestSuiteExecution, result *testkube.TestSuiteStepExecutionResult) {

	var testSuiteName string
	if testExecution.TestSuite != nil {
		testSuiteName = testExecution.TestSuite.Name
	}

	step := result.Step

	l := s.Log.With("type", step.Type(), "testSuiteName", testSuiteName, "name", step.FullName())

	switch step.Type() {

	case testkube.TestSuiteStepTypeExecuteTest:
		executeTestStep := step.Execute
		options, err := s.GetExecuteOptions(executeTestStep.Namespace, executeTestStep.Name, testkube.ExecutionRequest{
			Name:      fmt.Sprintf("%s-%s-%s", testSuiteName, executeTestStep.Name, rand.String(5)),
			Namespace: executeTestStep.Namespace,
			Params:    testExecution.Params,
		})

		if err != nil {
			result.Err(err)
			return
		}

		l.Debug("executing test", "params", testExecution.Params)
		options.Sync = true
		execution := s.executeTest(ctx, options)
		result.Execution = &execution

	case testkube.TestSuiteStepTypeDelay:
		l.Debug("delaying execution")
		time.Sleep(time.Millisecond * time.Duration(step.Delay.Duration))
		result.Execution.ExecutionResult.Success()

	default:
		result.Err(fmt.Errorf("can't find handler for execution step type: '%v'", step.Type()))
	}
}

func getExecutionsFilterFromRequest(c *fiber.Ctx) testresult.Filter {

	filter := testresult.NewExecutionsFilter()
	name := c.Query("id", "")
	if name != "" {
		filter = filter.WithName(name)
	}

	textSearch := c.Query("textSearch", "")
	if textSearch != "" {
		filter = filter.WithTextSearch(textSearch)
	}

	page, err := strconv.Atoi(c.Query("page", ""))
	if err == nil {
		filter = filter.WithPage(page)
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", ""))
	if err == nil && pageSize != 0 {
		filter = filter.WithPageSize(pageSize)
	}

	status := c.Query("status", "")
	if status != "" {
		filter = filter.WithStatus(testkube.ExecutionStatus(status))
	}

	dFilter := datefilter.NewDateFilter(c.Query("startDate", ""), c.Query("endDate", ""))
	if dFilter.IsStartValid {
		filter = filter.WithStartDate(dFilter.Start)
	}

	if dFilter.IsEndValid {
		filter = filter.WithEndDate(dFilter.End)
	}

	selector := c.Query("selector")
	if selector != "" {
		filter = filter.WithSelector(selector)
	}

	return filter
}

func mapToTestExecutionSummary(executions []testkube.TestSuiteExecution) []testkube.TestSuiteExecutionSummary {
	result := make([]testkube.TestSuiteExecutionSummary, len(executions))

	for i, execution := range executions {
		executionsSummary := make([]testkube.TestSuiteStepExecutionSummary, len(execution.StepResults))
		for j, stepResult := range execution.StepResults {
			executionsSummary[j] = mapStepResultToExecutionSummary(stepResult)
		}

		result[i] = testkube.TestSuiteExecutionSummary{
			Id:            execution.Id,
			Name:          execution.Name,
			TestSuiteName: execution.TestSuite.Name,
			Status:        execution.Status,
			StartTime:     execution.StartTime,
			EndTime:       execution.EndTime,
			Duration:      types.FormatDuration(execution.Duration),
			Execution:     executionsSummary,
		}
	}

	return result
}

func mapStepResultToExecutionSummary(r testkube.TestSuiteStepExecutionResult) testkube.TestSuiteStepExecutionSummary {
	var id, testName, name string
	var status *testkube.ExecutionStatus = testkube.ExecutionStatusPassed
	var stepType *testkube.TestSuiteStepType

	if r.Test != nil {
		testName = r.Test.Name
	}

	if r.Execution != nil {
		id = r.Execution.Id
		if r.Execution.ExecutionResult != nil {
			status = r.Execution.ExecutionResult.Status
		}
	}

	if r.Step != nil {
		stepType = r.Step.Type()
		name = r.Step.FullName()
	}

	return testkube.TestSuiteStepExecutionSummary{
		Id:       id,
		Name:     name,
		TestName: testName,
		Status:   status,
		Type_:    stepType,
	}
}

func mapTestSuiteUpsertRequestToTestCRD(request testkube.TestSuiteUpsertRequest) testsuitesv1.TestSuite {
	return testsuitesv1.TestSuite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
			Labels:    request.Labels,
		},
		Spec: testsuitesv1.TestSuiteSpec{
			Repeats:     int(request.Repeats),
			Description: request.Description,
			Before:      mapTestStepsToCRD(request.Before),
			Steps:       mapTestStepsToCRD(request.Steps),
			After:       mapTestStepsToCRD(request.After),
			Schedule:    request.Schedule,
		},
	}
}

func mapTestStepsToCRD(steps []testkube.TestSuiteStep) (out []testsuitesv1.TestSuiteStepSpec) {
	for _, step := range steps {
		out = append(out, mapTestStepToCRD(step))
	}

	return
}

func mapTestStepToCRD(step testkube.TestSuiteStep) (stepSpec testsuitesv1.TestSuiteStepSpec) {
	switch step.Type() {

	case testkube.TestSuiteStepTypeDelay:
		stepSpec.Delay = &testsuitesv1.TestSuiteStepDelay{
			Duration: step.Delay.Duration,
		}

	case testkube.TestSuiteStepTypeExecuteTest:
		s := step.Execute
		stepSpec.Execute = &testsuitesv1.TestSuiteStepExecute{
			Namespace: s.Namespace,
			Name:      s.Name,
			// TODO move StopOnFailure level up in operator model to mimic this one
			StopOnFailure: step.StopTestOnFailure,
		}
	}

	return
}
