package service

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

func CloseResponseBodyGracefully(httpResponse *http.Response) {
	if httpResponse == nil || httpResponse.Body == nil {
		return
	}
	err := httpResponse.Body.Close()
	if err != nil {
		common.SysError("failed to close response body: " + err.Error())
	}
}

func IOCopyBytesGracefully(c *gin.Context, src *http.Response, data []byte) {
	if c.Writer == nil {
		return
	}

	body := io.NopCloser(bytes.NewBuffer(data))

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	if src != nil {
		for k, v := range src.Header {
			// avoid setting Content-Length
			if k == "Content-Length" {
				continue
			}
			c.Writer.Header().Set(k, v[0])
		}
	}

	// set Content-Length header manually BEFORE calling WriteHeader
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Write header with status code (this sends the headers)
	if src != nil {
		c.Writer.WriteHeader(src.StatusCode)
	} else {
		c.Writer.WriteHeader(http.StatusOK)
	}

	_, err := io.Copy(c.Writer, body)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("failed to copy response body: %s", err.Error()))
	}
}

// DecompressResponseBody wraps the response body with appropriate decompressor
// based on Content-Encoding header. Supports gzip, brotli (br), and deflate.
// Returns the original body if no compression or unknown encoding.
func DecompressResponseBody(resp *http.Response) (io.ReadCloser, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}

	encoding := resp.Header.Get("Content-Encoding")
	common.SysLog(fmt.Sprintf("[DecompressResponseBody] Content-Encoding: %s, StatusCode: %d", encoding, resp.StatusCode))

	switch encoding {
	case "gzip":
		common.SysLog("[DecompressResponseBody] Using gzip decompression")
		return gzip.NewReader(resp.Body)
	case "br":
		common.SysLog("[DecompressResponseBody] Using brotli decompression")
		return io.NopCloser(brotli.NewReader(resp.Body)), nil
	case "deflate":
		common.SysLog("[DecompressResponseBody] Using deflate decompression")
		return flate.NewReader(resp.Body), nil
	default:
		common.SysLog(fmt.Sprintf("[DecompressResponseBody] No decompression needed (encoding: %s)", encoding))
		return resp.Body, nil
	}
}

// ReadResponseBody reads the entire response body with automatic decompression
// based on Content-Encoding header. Supports gzip, brotli (br), and deflate.
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	reader, err := DecompressResponseBody(resp)
	if err != nil {
		common.SysError(fmt.Sprintf("[ReadResponseBody] DecompressResponseBody error: %v", err))
		return nil, err
	}
	if reader == nil {
		return nil, nil
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		common.SysError(fmt.Sprintf("[ReadResponseBody] io.ReadAll error: %v", err))
		return nil, err
	}

	// 打印前100个字节用于调试
	if len(data) > 0 {
		previewLen := len(data)
		if previewLen > 100 {
			previewLen = 100
		}
		common.SysLog(fmt.Sprintf("[ReadResponseBody] Read %d bytes, first %d bytes (hex): %x", len(data), previewLen, data[:previewLen]))
		common.SysLog(fmt.Sprintf("[ReadResponseBody] First %d bytes (string): %s", previewLen, string(data[:previewLen])))
	}

	return data, nil
}
