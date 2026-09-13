package dto

// ListProductsResponse is the payload returned inside Response.Data for
// GET /api/v1/products. The endpoint takes no request body.
type ListProductsResponse struct {
	Items []ProductItem `json:"items"`
}

type ProductItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
}
