package services

import (
	"dgw-technical-test/internal/models/product"
	"dgw-technical-test/internal/repositories/product"
	"fmt"
)

type ProductService struct {
	ProductRepo *repositories.ProductRepository
}

func NewProductService(productRepo *repositories.ProductRepository) *ProductService {
	return &ProductService{ProductRepo: productRepo}
}

// GetAllProductsService retrieves all products from the database
func (s *ProductService) GetAllProductsService() ([]models.Product, error) {
	products, err := s.ProductRepo.GetAllProductsRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve seed products: %w", err)
	}
	return products, nil
}