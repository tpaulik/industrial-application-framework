package admission

import (
	"bytes"
	"errors"
	"github.com/onsi/ginkgo"
	"io"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
)

type AsyncAdmissionRequestSender struct {
	Url          string
	inProgress   bool
	errChannel   chan error
	httpResponse http.Response
}

func SendAdmissionReviewRequest(url string, admissionReviewReq v1beta1.AdmissionReview) (*v1beta1.AdmissionReview, error) {
	requestbody, err := json.Marshal(admissionReviewReq)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestbody))

	if err != nil {
		return &v1beta1.AdmissionReview{}, err
	} else if resp.StatusCode != 200 {
		return &v1beta1.AdmissionReview{}, errors.New("Response status code is: " + string(resp.StatusCode))
	}
	ar, err := parseResponseAsAdmissionReview(resp.Body)
	if err != nil {
		return &v1beta1.AdmissionReview{}, err
	}
	return ar, nil
}

// sending the request in a different go routine to avoid blocking of the main thread and to be able to react on the changes by the webhook
func (webhook *AsyncAdmissionRequestSender) SendAdmissionReviewRequestAsync(admissionReviewReq v1beta1.AdmissionReview) error {
	if webhook.inProgress {
		return errors.New("A previous request is in progress")
	}
	webhook.errChannel = make(chan error)
	webhook.inProgress = true
	go func() {
		defer ginkgo.GinkgoRecover()
		defer close(webhook.errChannel)

		requestbody, err := json.Marshal(admissionReviewReq)
		resp, err := http.Post(webhook.Url, "application/json", bytes.NewBuffer(requestbody))

		if err != nil {
			webhook.errChannel <- err
			return
		} else if resp.StatusCode != 200 {
			webhook.errChannel <- errors.New("Response status code is: " + string(resp.StatusCode))
			return
		}
		webhook.httpResponse = *resp
	}()
	return nil
}

func (webhook *AsyncAdmissionRequestSender) WaitAndReceiveResponseAsAdmissionReview() (v1beta1.AdmissionReview, error) {
	// wait for the webhook to finish and return response
	if webhook.errChannel == nil {
		return v1beta1.AdmissionReview{}, errors.New("Request was not initiated, cannot read response")
	}
	err := <-webhook.errChannel
	if err != nil {
		webhook.inProgress = false
		return v1beta1.AdmissionReview{}, err
	}
	ar, err := webhook.parseResponseAsAdmissionReview()
	webhook.inProgress = false
	if err != nil {
		return v1beta1.AdmissionReview{}, err
	}
	return *ar, nil
}

func (webhook *AsyncAdmissionRequestSender) parseResponseAsAdmissionReview() (*v1beta1.AdmissionReview, error) {
	defer webhook.httpResponse.Body.Close()
	return parseResponseAsAdmissionReview(webhook.httpResponse.Body)
}

func parseResponseAsAdmissionReview(httpBody io.ReadCloser) (*v1beta1.AdmissionReview, error) {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return &v1beta1.AdmissionReview{}, err
	}
	var ar *v1beta1.AdmissionReview
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return &v1beta1.AdmissionReview{}, err
	}
	return ar, nil
}
