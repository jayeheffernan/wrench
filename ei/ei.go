package ei

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/nightrune/wrench/logging"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

const EI_URL = "https://build.electricimp.com/v4/"
const BASE_EI_URL = "https://build.electricimp.com/"
const MODELS_ENDPOINT = "models"
const MODELS_REVISIONS_ENDPOINT = "revisions"
const DEVICES_ENDPOINT = "devices"
const DEVICES_LOG_ENDPOINT = "logs"
const MODELS_DEVICE_RESTART_ENDPOINT = "restart"

type DeviceListResponse struct {
	Success bool     `json:"success"`
	Devices []Device `json:"devices"`
}

type Device struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	ModelId     string `json:"model_id,omitempty"`
	PowerState  string `json:"powerstate,omitempty"`
	Rssi        int    `json:"rssi,omitempty"`
	AgentId     string `json:"agent_id,omitempty"`
	AgentStatus string `json:"agent_status,omitempty"`
}

type Model struct {
	Id      string   `json:"id,omitempty"`
	Name    string   `json:"name"`
	Devices []string `json:"devices,omitempty"`
}

type ModelList struct {
	Models []Model `json:"models"`
}

type ModelError struct {
	Code         string `json:"code"`
	MessageShort string `json:"message_short"`
	MessageFull  string `json:"message_full"`
}

type ModelResponse struct {
	Model   Model      `json:"model"`
	Success bool       `json:"success"`
	Error   ModelError `json:"error,omitempty"`
}

type ErrorDetails struct {
  Row int `json:"row"`
  Column int `json:"column"`
  Error string `json:"error"`
}

type BuildErrorDetails struct {
  DeviceErrors ErrorDetails `json:device_errors,omitempty`
  AgentErrors ErrorDetails `json:agent_errors,omitempty`
}

type BuildError struct {
	Code         string `json:"code"`
	ShortMessage string `json:"message_short"`
	FullMessage string  `json:"message_full"`
	Details BuildErrorDetails `json:"details"`
}

type CodeRevisionResponse struct {
	Success   bool             `json:"success"`
	Revisions CodeRevisionLong `json:"revision,omitempty"`
	Error     BuildError       `json:"error,omitempty"`
}

type CodeRevisionsResponse struct {
	Success   bool                `json:"success"`
	Revisions []CodeRevisionShort `json:"revisions"`
}

type CodeRevisionShort struct {
	Version      int    `json:"version"`
	CreatedAt    string `json:"created_at"`
	ReleaseNotes string `json:"release_notes"`
}

type CodeRevisionLong struct {
	Version      int    `json:"version,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	DeviceCode   string `json:"device_code,omitempty"`
	AgentCode    string `json:"agent_code,omitempty"`
	ReleaseNotes string `json:"release_notes,omitempty"`
}

type BuildClient struct {
	creds       string
	http_client *http.Client
}

type DeviceResponse struct {
  Success bool   `json:"success"`
  Device  Device `json:"device"`
}

type DeviceLogEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

type DeviceLogResponse struct {
	Logs    []DeviceLogEntry `json:"logs"`
	PollUrl string           `json:"poll_url"`
	Success bool             `json:"success"`
}

func NewBuildClient(api_key string) *BuildClient {
	client := new(BuildClient)
	client.http_client = &http.Client{}
	cred_data := []byte(api_key)
	client.creds = base64.StdEncoding.EncodeToString(cred_data)
	return client
}

func Concat(a string, b string) string {
	var buffer bytes.Buffer
	buffer.WriteString(a)
	buffer.WriteString(b)
	return buffer.String()
}

func (m BuildClient) SetAuthHeader(request *http.Request) {
	request.Header.Set("Authorization", "Basic "+m.creds)
}

type Timeout struct {

}

func (m Timeout) Error() string {
  return "Timed out"
}

func (m *BuildClient) _complete_request(method string,
	url string, data []byte) ([]byte, error) {
	var req *http.Request
	if data != nil {
		req, _ = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	m.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.http_client.Do(req)
  if resp.StatusCode == http.StatusGatewayTimeout {
    return nil, new(Timeout)
  }

	if err == nil {
		dump, err := httputil.DumpResponse(resp, true)
		logging.Debug(string(dump))
		full_response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return full_response, err
		}
		return full_response, nil
	} else {
		return nil, err
	}
}

func (m *BuildClient) ListModels() (*ModelList, error) {
	list := new(ModelList)
	full_resp, err := m._complete_request("GET", Concat(EI_URL, "models"), nil)
	if err != nil {
		logging.Debug("An error happened during model get, %s", err.Error())
		return list, err
	}

	if err := json.Unmarshal(full_resp, list); err != nil {
		logging.Warn("Failed to unmarshal data from models.. %s", err.Error())
		return list, err
	}

	return list, nil
}

func (m *BuildClient) CreateModel(new_model *Model) (*Model, error) {
	var url bytes.Buffer
	resp := new(ModelResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)

	req_string, err := json.Marshal(new_model)
	logging.Debug("Request String for upload: %s", req_string)
	full_resp, err := m._complete_request("POST", url.String(), req_string)
	if err != nil {
		logging.Debug("An error happened during model creation, %s", err.Error())
		return &resp.Model, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from model response.. %s", err.Error())
		return &resp.Model, err
	}

	return &resp.Model, nil
}

func (m *BuildClient) UpdateModel(model_id string, new_model *Model) (*Model, error) {
	var url bytes.Buffer
	resp := new(ModelResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)

	req_string, err := json.Marshal(new_model)
	logging.Debug("Request String for upload: %s", req_string)
	full_resp, err := m._complete_request("PUT", url.String(), req_string)
	if err != nil {
		logging.Debug("An error happened during model creation, %s", err.Error())
		return &resp.Model, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from model response.. %s", err.Error())
		return &resp.Model, err
	}

	return &resp.Model, nil
}

func (m *BuildClient) GetModel(model_id string) (*Model, error) {
	var url bytes.Buffer
	resp := new(ModelResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)

	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("An error happened during model get, %s", err.Error())
		return &resp.Model, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from get model response.. %s", err.Error())
		return &resp.Model, err
	}

	if resp.Success == false {
		return &resp.Model, errors.New("Error when attempting to get model: " + resp.Error.MessageShort)
	}

	return &resp.Model, nil
}

func (m *BuildClient) DeleteModel(model_id string) error {
	var url bytes.Buffer
	resp := new(ModelResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)

	full_resp, err := m._complete_request("DELETE", url.String(), nil)
	if err != nil {
		logging.Debug("An error happened during model deletion, %s", err.Error())
		return err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from model response.. %s", err.Error())
		return err
	}

	if resp.Success == false {
		return errors.New("Error When retriveing Code Revisions")
	}

	return nil
}

func (m *BuildClient) RestartModelDevices(model_id string) error {
	var url bytes.Buffer
	resp := new(ModelResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)
	url.WriteString("/")
	url.WriteString(MODELS_DEVICE_RESTART_ENDPOINT)

	full_resp, err := m._complete_request("POST", url.String(), nil)
	if err != nil {
		logging.Debug("An error happened during model restart, %s", err.Error())
		return err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from model response.. %s", err.Error())
		return err
	}

	if resp.Success == false {
		return errors.New("Error When retriveing Code Revisions")
	}

	return nil
}

func (m *BuildClient) GetCodeRevisionList(model_id string) (
	[]CodeRevisionShort, error) {
	var url bytes.Buffer
	resp := new(CodeRevisionsResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)
	url.WriteString("/")
	url.WriteString(MODELS_REVISIONS_ENDPOINT)
	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get code revisions: %s", err.Error())
		return resp.Revisions, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision.. %s", err.Error())
		return resp.Revisions, err
	}

	if resp.Success == false {
		return resp.Revisions, errors.New("Error When retriveing Code Revisions")
	}
	return resp.Revisions, nil
}

func (m *BuildClient) GetCodeRevision(model_id string, build_num string) (CodeRevisionLong, error) {
	var url bytes.Buffer
	resp := new(CodeRevisionResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)
	url.WriteString("/")
	url.WriteString(MODELS_REVISIONS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(build_num)
	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get code revisions: %s", err.Error())
		return resp.Revisions, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision.. %s", err.Error())
		return resp.Revisions, err
	}

	if resp.Success == false {
		return resp.Revisions, errors.New("Error When retriveing Code Revisions")
	}
	return resp.Revisions, nil
}

func (m *BuildClient) UpdateCodeRevision(model_id string,
	request *CodeRevisionLong) (*CodeRevisionResponse, error) {
	var url bytes.Buffer
	resp := new(CodeRevisionResponse)
	url.WriteString(EI_URL)
	url.WriteString(MODELS_ENDPOINT)
	url.WriteString("/")
	url.WriteString(model_id)
	url.WriteString("/")
	url.WriteString(MODELS_REVISIONS_ENDPOINT)

	req_string, err := json.Marshal(request)
	logging.Debug("Request String for upload: %s", req_string)
	full_resp, err := m._complete_request("POST", url.String(), req_string)
	if err != nil {
		logging.Debug("Failed to update code revisions: %s", err.Error())
		return nil, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision update.. %s", err.Error())
		return nil, err
	}

	return resp, nil
}

func (m *BuildClient) GetDeviceList() ([]Device, error) {
	var url bytes.Buffer
	resp := new(DeviceListResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)

	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get device list: %s", err.Error())
		return resp.Devices, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision update.. %s", err.Error())
		return resp.Devices, err
	}

	if resp.Success == false {
		return resp.Devices, errors.New("Error When retriveing Code Revisions")
	}
	return resp.Devices, nil
}

func (m *BuildClient) GetDeviceLogs(device_id string) ([]DeviceLogEntry, string, error) {
	var url bytes.Buffer
	resp := new(DeviceLogResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)
	url.WriteString("/")
	url.WriteString(device_id)
	url.WriteString("/")
	url.WriteString(DEVICES_LOG_ENDPOINT)
	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get device logs: %s", err.Error())
		return resp.Logs, "", err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from device logs.. %s", err.Error())
		return resp.Logs, "", err
	}

	if resp.Success == false {
		return resp.Logs, "", errors.New("Error When retriveing device logs")
	}
	return resp.Logs, resp.PollUrl, nil
}

func (m *BuildClient) ContinueDeviceLogs(poll_url string) ([]DeviceLogEntry, error) {
  var url bytes.Buffer
  resp := new(DeviceLogResponse)
  url.WriteString(BASE_EI_URL)
  url.WriteString(poll_url)
  full_resp, err := m._complete_request("GET", url.String(), nil)
  if err != nil {
    logging.Debug("Failed to get device logs: %s", err.Error())
    return resp.Logs, err
  }

  if err := json.Unmarshal(full_resp, resp); err != nil {
    logging.Warn("Failed to unmarshal data from device logs.. %s", err.Error())
    return resp.Logs, err
  }

  if resp.Success == false {
    return resp.Logs, errors.New("Error when retriveing device logs")
  }
  return resp.Logs, nil
}

func (m *BuildClient) GetDevice(device_id string) (Device, error) {
	var url bytes.Buffer
	resp := new(DeviceResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)
	url.WriteString("/")
	url.WriteString(device_id)

	full_resp, err := m._complete_request("GET", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get device list: %s", err.Error())
		return resp.Device, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision update.. %s", err.Error())
		return resp.Device, err
	}

	if resp.Success == false {
		return resp.Device, errors.New("Error When retriveing Code Revisions")
	}
	return resp.Device, nil
}

const DEVICES_RESTART_ENDPOINT = "restart"

func (m *BuildClient) RestartDevice(device_id string) error {
	var url bytes.Buffer
	resp := new(DeviceResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)
	url.WriteString("/")
	url.WriteString(device_id)
	url.WriteString("/")
	url.WriteString(DEVICES_RESTART_ENDPOINT)

	full_resp, err := m._complete_request("POST", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to get device list: %s", err.Error())
		return err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from code revision update.. %s", err.Error())
		return err
	}

	if resp.Success == false {
		return errors.New("Error When retriveing Code Revisions")
	}
	return nil
}

func (m *BuildClient) UpdateDevice(new_device *Device, device_id string) (Device, error) {
	var url bytes.Buffer
	resp := new(DeviceResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)
	url.WriteString("/")
	url.WriteString(device_id)

	req_bytes, err := json.Marshal(new_device)
	full_resp, err := m._complete_request("PUT", url.String(), req_bytes)
	if err != nil {
		logging.Debug("Failed to update device: %s", err.Error())
		return resp.Device, err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from device update.. %s", err.Error())
		return resp.Device, err
	}

	if resp.Success == false {
		return resp.Device, errors.New("Error when updating device")
	}
	return resp.Device, nil
}

func (m *BuildClient) DeleteDevice(device_id string) error {
	var url bytes.Buffer
	resp := new(DeviceResponse)
	url.WriteString(EI_URL)
	url.WriteString(DEVICES_ENDPOINT)
	url.WriteString("/")
	url.WriteString(device_id)

	full_resp, err := m._complete_request("DELETE", url.String(), nil)
	if err != nil {
		logging.Debug("Failed to delete device: %s", err.Error())
		return err
	}

	if err := json.Unmarshal(full_resp, resp); err != nil {
		logging.Warn("Failed to unmarshal data from device deletion.. %s", err.Error())
		return err
	}

	if resp.Success == false {
		return errors.New("Error when updating device")
	}
	return nil
}
