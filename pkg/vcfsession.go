package vcf

import (
	"bytes"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	sl "github.com/eshu0/simplelogger/interfaces"
)

type VCFSession struct {
	FQDN               string           `json:"fqdn"`
	Base64AuthInfo     string           `json:"base64authinfo"`
	Logger             sl.ISimpleLogger `json:"-"`
	InsecureSkipVerify bool             `json:"insecureskipverify"`
}

func NewVCFSession(FQDN string, Logger sl.ISimpleLogger) VCFSession {
	sess := VCFSession{}
	sess.FQDN = FQDN
	sess.Logger = Logger
	return sess
}

func (vmcfs *VCFSession) LogMessage(cmd string, Message string) {
	fmt.Println(Message)
	vmcfs.Logger.LogInfo(cmd, Message)
}

func createHeaders(base64AuthInfo string) http.Header {

	headers := http.Header{}

	headers.Add("Accept", "application/json")

	if base64AuthInfo != "" {
		base64AuthInfo := fmt.Sprintf("Basic %s", base64AuthInfo)
		headers.Add("Authorization", base64AuthInfo)
	}

	return headers
}

func (vmcfs *VCFSession) SendRequest(Resource string, ContentType string, methodin string, tosend io.Reader) (*http.Response, bool, error) {

	url := fmt.Sprintf("https://%s/v1/%s", vmcfs.FQDN, Resource)
	req := &http.Request{}

	if tosend != nil {
		req, _ = http.NewRequest(methodin, url, tosend)
	} else {
		req, _ = http.NewRequest(methodin, url, nil)
	}

	req.Header = createHeaders(vmcfs.Base64AuthInfo)

	if vmcfs.InsecureSkipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		vmcfs.LogMessage("SendRequest", fmt.Sprintf("Failed to connect with error %s\n", err))
		return nil, false, err
	}

	if res == nil {
		vmcfs.LogMessage("SendRequest", fmt.Sprintf("Failed to connect with error result was empty"))
		return nil, false, errors.New("Failed to connect with error result was empty")
	} else {
		return res, true, nil
	}

}

func (vmcfs *VCFSession) GETResourceRequest(Resource string) (*http.Response, bool, error) {
	return vmcfs.SendRequest(Resource, "application/json", "GET", nil)
}

func (vmcfs *VCFSession) UploadFile(Filepath string, Resource string, MethodIn string) (*http.Response, bool, error) {

	// New multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Cert Files.
	mediaData, _ := ioutil.ReadFile(Filepath)
	mediaHeader := textproto.MIMEHeader{}
	mediaHeader.Set("Content-Type", "application/octet-stream")
	mediaHeader.Set("Content-Disposition", fmt.Sprintf("form-data; name=file; filename=\"%v\".", Filepath))
	mediaHeader.Set("Content-ID", "media")
	mediaHeader.Set("Content-Filename", Filepath)

	mediaPart, _ := writer.CreatePart(mediaHeader)
	io.Copy(mediaPart, bytes.NewReader(mediaData))

	// Close multipart writer.
	writer.Close()

	return vmcfs.SendRequest(Resource, fmt.Sprintf("multipart/related; boundary=%s", writer.Boundary()), MethodIn, bytes.NewReader(body.Bytes()))
}

func (vmcfs *VCFSession) BuildAuth(Username string, Password string) {
	AuthInfo := fmt.Sprintf("%s:%s", Username, Password)
	sEnc := b64.StdEncoding.EncodeToString([]byte(AuthInfo))
	vmcfs.Base64AuthInfo = sEnc
}

func (vmcfs *VCFSession) Save(FilePath string, Log sl.ISimpleLogger) bool {
	bytes, err1 := json.MarshalIndent(vmcfs, "", "\t") //json.Marshal(p)
	if err1 != nil {
		Log.LogErrorf("SaveToFile()", "Marshal json for %s failed with %s ", FilePath, err1.Error())
		return false
	}

	err2 := ioutil.WriteFile(FilePath, bytes, 0644)
	if err2 != nil {
		Log.LogErrorf("SaveToFile()", "Saving %s failed with %s ", FilePath, err2.Error())
		return false
	}

	return true

}

func (vmcfs *VCFSession) Load(FilePath string, Log sl.ISimpleLogger) (*VCFSession, bool) {
	ok, err := vmcfs.checkFileExists(FilePath)
	if ok {
		bytes, err1 := ioutil.ReadFile(FilePath) //ReadAll(jsonFile)
		if err1 != nil {
			Log.LogErrorf("LoadFile()", "Reading '%s' failed with %s ", FilePath, err1.Error())
			return nil, false
		}

		vcfs := VCFSession{}

		err2 := json.Unmarshal(bytes, &vcfs)

		if err2 != nil {
			Log.LogErrorf("LoadFile()", " Loading %s failed with %s ", FilePath, err2.Error())
			return nil, false
		}

		Log.LogDebugf("LoadFile()", "Read Base64AuthInfo %s ", vcfs.Base64AuthInfo)
		Log.LogDebugf("LoadFile()", "Read FQDN %s ", vcfs.FQDN)

		return &vcfs, true
	} else {

		if err != nil {
			Log.LogErrorf("LoadFile()", "'%s' was not found to load with error: %s", FilePath, err.Error())
		} else {
			Log.LogErrorf("LoadFile()", "'%s' was not found to load", FilePath)
		}

		return nil, false
	}
}

func (vmcfs *VCFSession) checkFileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, err
	}
	return !info.IsDir(), nil
}
