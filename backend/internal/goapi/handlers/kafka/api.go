package kafka

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TestSaleInvoiceConsumer - API endpoint to test the sale invoice consumer
func TestSaleInvoiceConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessageSaleInvoiceCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Sale invoice consumer test completed successfully",
	})
}

// TestSaleReturnConsumer - API endpoint to test the sale return consumer
func TestSaleReturnConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessageSaleInvoiceReturnCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Sale return consumer test completed successfully",
	})
}

// TestInventoryConsumer - API endpoint to test the inventory consumer
func TestInventoryConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessageInventoryCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Inventory consumer test completed successfully",
	})
}

// TestWarehouseConsumer - API endpoint to test the warehouse consumer
func TestWarehouseConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessageWarehouseCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Warehouse consumer test completed successfully",
	})
}

// TestSaleOrderConsumer - API endpoint to test the sale order consumer
func TestSaleOrderConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessageSaleOrderCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Sale order consumer test completed successfully",
	})
}

// TestPurchaseConsumer - API endpoint to test the purchase consumer
func TestPurchaseConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessagePurchaseCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase consumer test completed successfully",
	})
}

// TestPurchaseOrderConsumer - API endpoint to test the purchase order consumer
func TestPurchaseOrderConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessagePurchaseOrderCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase order consumer test completed successfully",
	})
}

// TestPurchasePartialConsumer - API endpoint to test the purchase partial consumer
func TestPurchasePartialConsumer(c echo.Context) error {
	// Get JSON payload from request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Convert payload back to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to marshal JSON",
		})
	}

	// Test the consumer function
	err = OnConsumeMessagePurchasePartialCreateOrUpdate(string(jsonBytes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase partial consumer test completed successfully",
	})
}

// GetConsumerStatus - API endpoint to get the status of all consumers
func GetConsumerStatus(c echo.Context) error {
	config := NewKafkaConfig()

	status := map[string]interface{}{
		"kafka_server":    config.ServerURL,
		"is_valid":        config.IsValid(),
		"batch_size":      config.BatchSize,
		"consumer_groups": config.ConsumerGroups,
		"topics": map[string]interface{}{
			"sale_invoice": []string{
				TOPIC_SALE_INVOICE_CREATE,
				TOPIC_SALE_INVOICE_UPDATE,
				TOPIC_SALE_INVOICE_DELETE,
			}, "sale_return": []string{
				TOPIC_SALE_INVOICE_RETURN_CREATE,
				TOPIC_SALE_INVOICE_RETURN_UPDATE,
				TOPIC_SALE_INVOICE_RETURN_DELETE,
			},
			"sale_order": []string{
				TOPIC_SALE_ORDER_CREATE,
				TOPIC_SALE_ORDER_UPDATE,
				TOPIC_SALE_ORDER_DELETE,
			},
			"inventory": []string{
				TOPIC_INVENTORY_CREATE,
				TOPIC_INVENTORY_UPDATE,
				TOPIC_INVENTORY_DELETE,
				MQ_TOPIC_BULK_CREATED,
				MQ_TOPIC_BULK_UPDATED,
				MQ_TOPIC_BULK_DELETED,
			},
			"warehouse": []string{
				TOPIC_WAREHOUSE_CREATE,
				TOPIC_WAREHOUSE_UPDATE,
				TOPIC_WAREHOUSE_DELETE,
			},
			"purchase": []string{
				TOPIC_PURCHASE_ORDER_CREATE,
				TOPIC_PURCHASE_ORDER_UPDATE,
				TOPIC_PURCHASE_ORDER_DELETE,
				TOPIC_PURCHASE_CREATE,
				TOPIC_PURCHASE_UPDATE,
				TOPIC_PURCHASE_DELETE,
				TOPIC_PURCHASE_RECEIVE_CREATE,
				TOPIC_PURCHASE_RECEIVE_UPDATE,
				TOPIC_PURCHASE_RECEIVE_DELETE,
			},
		},
	}

	return c.JSON(http.StatusOK, status)
}
