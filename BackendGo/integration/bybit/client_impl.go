package bybit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// client реализация клиента Bybit
type client struct {
	baseURL    string
	recvWindow int
	httpClient *http.Client
}

// NewClient создает новый клиент Bybit
func NewClient(baseURL string, recvWindow int) Client {
	return &client{
		baseURL:    baseURL,
		recvWindow: recvWindow,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWalletBalance получает баланс кошелька
func (c *client) GetWalletBalance(ctx context.Context, account *BybitAccount) (*BybitWalletBalance, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	queryParams := fmt.Sprintf("accountType=%s", account.AccountType)

	signature := c.generateSignature(timestamp, queryParams, account)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/account/wallet-balance?%s", c.baseURL, queryParams), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	// Логируем ответ для отладки
	resultBytes, err := json.Marshal(bybitResp.Result)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга результата: %w", err)
	}
	log.Printf("Bybit API Response: %s", string(resultBytes))

	// Преобразуем result в map[string]interface{}
	resultMap, ok := bybitResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неожиданный формат ответа: %v", bybitResp.Result)
	}

	// Преобразуем map обратно в JSON
	resultBytes, err = json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга результата: %w", err)
	}

	var result BybitWalletBalance
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetInstruments получает список доступных для торговли пар
func (c *client) GetInstruments(ctx context.Context, category string) (*BybitInstrumentsResponse, error) {
	queryParams := url.Values{}
	queryParams.Set("category", category)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/market/instruments-info?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	// Преобразуем result в map[string]interface{}
	resultMap, ok := bybitResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неожиданный формат ответа: %v", bybitResp.Result)
	}

	// Преобразуем map обратно в JSON
	resultBytes, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга результата: %w", err)
	}

	var result BybitInstrumentsResponse
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetTickers получает текущие котировки
func (c *client) GetTickers(ctx context.Context, category string, symbol *string) (*BybitTickersResponse, error) {
	queryParams := url.Values{}
	queryParams.Set("category", category)
	if symbol != nil {
		queryParams.Set("symbol", *symbol)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/market/tickers?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitTickersResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetKlines получает исторические свечи
func (c *client) GetKlines(
	ctx context.Context,
	category string,
	symbol string,
	interval string,
	limit int,
	start *time.Time,
	end *time.Time,
) (*BybitKlinesResponse, error) {
	queryParams := url.Values{}
	queryParams.Set("category", category)
	queryParams.Set("symbol", symbol)
	queryParams.Set("interval", interval)
	queryParams.Set("limit", strconv.Itoa(limit))

	if start != nil {
		queryParams.Set("start", strconv.FormatInt(start.UnixMilli(), 10))
	}
	if end != nil {
		queryParams.Set("end", strconv.FormatInt(end.UnixMilli(), 10))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/market/kline?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitKlinesResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetTrades получает исторические сделки
func (c *client) GetTrades(
	ctx context.Context,
	category string,
	symbol string,
	limit int,
	orderID *string,
) (*BybitTradesResponse, error) {
	queryParams := url.Values{}
	queryParams.Set("category", category)
	queryParams.Set("symbol", symbol)
	queryParams.Set("limit", strconv.Itoa(limit))

	if orderID != nil {
		queryParams.Set("orderId", *orderID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/market/recent-trade?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitTradesResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// CreateOrder создает ордер
func (c *client) CreateOrder(
	ctx context.Context,
	account *BybitAccount,
	symbol string,
	side string,
	orderType string,
	qty string,
	price *string,
	timeInForce string,
	orderLinkID *string,
) (*BybitOrderResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	payload := map[string]interface{}{
		"category":    "spot",
		"symbol":      symbol,
		"side":        side,
		"orderType":   orderType,
		"qty":         qty,
		"timeInForce": timeInForce,
	}

	if price != nil {
		payload["price"] = *price
	}
	if orderLinkID != nil {
		payload["orderLinkId"] = *orderLinkID
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	signature := c.generateSignature(timestamp, string(payloadBytes), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v5/order/create", c.baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitOrderResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// AmendOrder изменяет ордер
func (c *client) AmendOrder(
	ctx context.Context,
	account *BybitAccount,
	symbol string,
	orderID string,
	price *string,
	qty *string,
) (*BybitOrderResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	payload := map[string]interface{}{
		"category": "spot",
		"symbol":   symbol,
		"orderId":  orderID,
	}

	if price != nil {
		payload["price"] = *price
	}
	if qty != nil {
		payload["qty"] = *qty
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	signature := c.generateSignature(timestamp, string(payloadBytes), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v5/order/amend", c.baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitOrderResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// CancelOrder отменяет ордер
func (c *client) CancelOrder(
	ctx context.Context,
	account *BybitAccount,
	symbol string,
	orderID string,
) (*BybitOrderResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	payload := map[string]interface{}{
		"category": "spot",
		"symbol":   symbol,
		"orderId":  orderID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	signature := c.generateSignature(timestamp, string(payloadBytes), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v5/order/cancel", c.baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitOrderResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// CancelAllOrders отменяет все ордера
func (c *client) CancelAllOrders(
	ctx context.Context,
	account *BybitAccount,
	symbol string,
) (*BybitOrderResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	payload := map[string]interface{}{
		"category": "spot",
		"symbol":   symbol,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	signature := c.generateSignature(timestamp, string(payloadBytes), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v5/order/cancel-all", c.baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitOrderResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetOpenOrders получает открытые ордера
func (c *client) GetOpenOrders(
	ctx context.Context,
	account *BybitAccount,
	symbol string,
	orderID *string,
	limit int,
) (*BybitOrderListResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	queryParams := url.Values{}
	queryParams.Set("category", "spot")
	queryParams.Set("symbol", symbol)
	if orderID != nil {
		queryParams.Set("orderId", *orderID)
	}
	queryParams.Set("limit", strconv.Itoa(limit))

	signature := c.generateSignature(timestamp, queryParams.Encode(), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/order/realtime?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	var result BybitOrderListResponse
	if err := json.Unmarshal(bybitResp.Result.(json.RawMessage), &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// GetFeeRate получает ставки комиссии
func (c *client) GetFeeRate(
	ctx context.Context,
	account *BybitAccount,
	category string,
	symbol *string,
	baseCoin *string,
) (*BybitFeeRateResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	queryParams := url.Values{}
	queryParams.Set("category", category)
	if symbol != nil {
		queryParams.Set("symbol", *symbol)
	}
	if baseCoin != nil {
		queryParams.Set("baseCoin", *baseCoin)
	}

	signature := c.generateSignature(timestamp, queryParams.Encode(), account)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v5/account/fee-rate?%s", c.baseURL, queryParams.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-BAPI-API-KEY", account.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	var bybitResp BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&bybitResp); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if !bybitResp.IsSuccess() {
		return nil, fmt.Errorf("ошибка API: %s", bybitResp.RetMsg)
	}

	// Логируем ответ для отладки
	resultBytes, err := json.Marshal(bybitResp.Result)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга результата: %w", err)
	}
	log.Printf("Bybit API Response: %s", string(resultBytes))

	// Преобразуем result в map[string]interface{}
	resultMap, ok := bybitResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неожиданный формат ответа: %v", bybitResp.Result)
	}

	// Преобразуем map обратно в JSON
	resultBytes, err = json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга результата: %w", err)
	}

	var result BybitFeeRateResponse
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результата: %w", err)
	}

	return &result, nil
}

// generateSignature генерирует подпись для запроса
func (c *client) generateSignature(timestamp string, queryParams string, account *BybitAccount) string {
	paramStr := timestamp + account.APIKey + strconv.Itoa(c.recvWindow) + queryParams
	h := hmac.New(sha256.New, []byte(account.APISecret))
	h.Write([]byte(paramStr))
	return hex.EncodeToString(h.Sum(nil))
} 