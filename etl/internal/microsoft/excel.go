package microsoft

import (
	"context"
	"fmt"
	"net/url"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// Client wraps the Microsoft Graph SDK for Excel operations
type Client struct {
	serviceClient *msgraphsdk.GraphServiceClient
	filePath      string // Path to the Excel file in OneDrive/SharePoint
	sheetName     string
	driveId       string // Cached drive ID
}

// NewClient creates a new Microsoft Graph client for Excel file access
func NewClient(ctx context.Context, credential azcore.TokenCredential, filePath string) (*Client, error) {
	if credential == nil {
		return nil, fmt.Errorf("microsoft: credential is required")
	}
	if filePath == "" {
		return nil, fmt.Errorf("microsoft: file path is required")
	}

	// Define scopes needed for Excel file access
	scopes := []string{
		"https://graph.microsoft.com/Files.Read.All",
	}

	// Create Graph service client
	serviceClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(credential, scopes)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to create graph service client: %w", err)
	}

	// Resolve the drive ID upfront
	drive, err := serviceClient.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to resolve drive: %w", err)
	}
	if drive.GetId() == nil {
		return nil, fmt.Errorf("microsoft: drive ID is nil")
	}

	return &Client{
		serviceClient: serviceClient,
		filePath:      filePath,
		sheetName:     "TARGETS",
		driveId:       *drive.GetId(),
	}, nil
}

// RowData represents a row of data from Excel
type RowData struct {
	Values []string
}

// fileItem returns the DriveItem for the configured file path.
// Uses the Graph API path syntax: /drives/{id}/root:/{path}
func (c *Client) fileItem(ctx context.Context) (models.DriveItemable, error) {
	encodedPath := url.PathEscape(c.filePath)
	fileURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/drives/%s/root:/%s", c.driveId, encodedPath)
	item, err := c.serviceClient.Drives().ByDriveId(c.driveId).Root().WithUrl(fileURL).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to get file '%s': %w", c.filePath, err)
	}
	return item, nil
}

// driveItems returns the Items request builder for the drive
func (c *Client) driveItems() *drives.ItemItemsRequestBuilder {
	return c.serviceClient.Drives().ByDriveId(c.driveId).Items()
}

// ReadSheet reads data from an Excel worksheet
func (c *Client) ReadSheet(ctx context.Context, sheetName string, startRow int, endRow int) ([]RowData, error) {
	if sheetName == "" {
		sheetName = c.sheetName
	}

	fileItem, err := c.fileItem(ctx)
	if err != nil {
		return nil, err
	}
	if fileItem.GetId() == nil {
		return nil, fmt.Errorf("microsoft: file item ID is nil")
	}

	// Get the worksheet by name (filter the collection)
	worksheets, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to list worksheets: %w", err)
	}

	var sheetID string
	for _, ws := range worksheets.GetValue() {
		if ws.GetName() != nil && *ws.GetName() == sheetName {
			if ws.GetId() != nil {
				sheetID = *ws.GetId()
			}
			break
		}
	}
	if sheetID == "" {
		return nil, fmt.Errorf("microsoft: worksheet '%s' not found", sheetName)
	}

	// Get used range to determine actual dimensions
	usedRange, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().ByWorkbookWorksheetId(sheetID).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to get used range: %w", err)
	}

	// Calculate actual end row based on used range or provided endRow
	actualEndRow := endRow
	if usedRange.GetRowCount() != nil && *usedRange.GetRowCount() < int32(endRow) {
		actualEndRow = int(*usedRange.GetRowCount())
	}

	// Get range values
	rangeAddress := fmt.Sprintf("%s!A%d:Z%d", sheetName, startRow, actualEndRow)
	rangeObj, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().ByWorkbookWorksheetId(sheetID).RangeWithAddress(&rangeAddress).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to get range %s: %w", rangeAddress, err)
	}

	// Process the values
	var result []RowData
	untyped := rangeObj.GetValues().(*serialization.UntypedNode)
	rawValues := untyped.GetValue()
	values, ok := rawValues.([][]interface{})
	if !ok || len(values) == 0 {
		return nil, fmt.Errorf("microsoft: no data found in range %s", rangeAddress)
	}

	for _, row := range values {
		if len(row) == 0 {
			continue
		}
		valuesStr := make([]string, len(row))
		for i, v := range row {
			valuesStr[i] = fmt.Sprintf("%v", v)
		}
		result = append(result, RowData{Values: valuesStr})
	}

	return result, nil
}

// GetSheetTitle gets the header row from a worksheet
func (c *Client) GetSheetTitle(ctx context.Context, sheetName string) ([]string, error) {
	if sheetName == "" {
		sheetName = c.sheetName
	}

	fileItem, err := c.fileItem(ctx)
	if err != nil {
		return nil, err
	}
	if fileItem.GetId() == nil {
		return nil, fmt.Errorf("microsoft: file item ID is nil")
	}

	// Find the worksheet by name
	worksheets, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to list worksheets: %w", err)
	}

	var sheetID string
	for _, ws := range worksheets.GetValue() {
		if ws.GetName() != nil && *ws.GetName() == sheetName {
			if ws.GetId() != nil {
				sheetID = *ws.GetId()
			}
			break
		}
	}
	if sheetID == "" {
		return nil, fmt.Errorf("microsoft: worksheet '%s' not found", sheetName)
	}

	// Get first row (headers)
	rangeAddress := fmt.Sprintf("%s!1:1", sheetName)
	rangeObj, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().ByWorkbookWorksheetId(sheetID).RangeWithAddress(&rangeAddress).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to get headers: %w", err)
	}

	untyped := rangeObj.GetValues().(*serialization.UntypedNode)
	rawValues := untyped.GetValue()
	values, ok := rawValues.([][]interface{})
	if !ok || len(values) == 0 || len(values[0]) == 0 {
		return nil, fmt.Errorf("microsoft: no headers found in sheet %s", sheetName)
	}

	headers := make([]string, len(values[0]))
	for i, v := range values[0] {
		headers[i] = fmt.Sprintf("%v", v)
	}

	return headers, nil
}

// ListSheets gets all worksheet names in the workbook
func (c *Client) ListSheets(ctx context.Context) ([]string, error) {
	fileItem, err := c.fileItem(ctx)
	if err != nil {
		return nil, err
	}
	if fileItem.GetId() == nil {
		return nil, fmt.Errorf("microsoft: file item ID is nil")
	}

	worksheets, err := c.driveItems().ByDriveItemId(*fileItem.GetId()).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("microsoft: failed to list sheets: %w", err)
	}

	sheets := make([]string, 0, len(worksheets.GetValue()))
	for _, ws := range worksheets.GetValue() {
		if ws.GetName() != nil {
			sheets = append(sheets, *ws.GetName())
		}
	}

	return sheets, nil
}
