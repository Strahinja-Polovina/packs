package service

import (
	"context"
	"sort"

	"github.com/Strahinja-Polovina/packs/internal/domain/entity"
	"github.com/Strahinja-Polovina/packs/internal/domain/repository"
	"github.com/Strahinja-Polovina/packs/pkg/logger"
)

// PackService handles pack-related business logic
type PackService struct {
	packRepo repository.PackRepository
	logger   *logger.Logger
}

// NewPackService creates a new pack service
func NewPackService(packRepo repository.PackRepository, logger *logger.Logger) *PackService {
	return &PackService{
		packRepo: packRepo,
		logger:   logger,
	}
}

// PackCalculationRequest represents a request to calculate pack combinations
type PackCalculationRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

// PackCalculationResponse represents the response with calculated pack combinations
type PackCalculationResponse struct {
	Amount      int         `json:"amount"`
	PackSizes   []int       `json:"pack_sizes"`
	Combination map[int]int `json:"combination"`
	TotalPacks  int         `json:"total_packs"`
	TotalAmount int         `json:"total_amount"`
}

// CalculateOptimalPacks calculates the optimal pack combination for a given amount
func (s *PackService) CalculateOptimalPacks(ctx context.Context, req PackCalculationRequest) (*PackCalculationResponse, error) {
	s.logger.Info("Calculating optimal packs for amount: %d", req.Amount)

	if req.Amount <= 0 {
		s.logger.Error("Invalid amount provided: %d", req.Amount)
		return nil, entity.ErrInvalidAmount
	}

	packs := s.packRepo.List(ctx)
	packSizes := make([]int, len(packs))
	for i, pack := range packs {
		packSizes[i] = pack.Size()
	}

	if len(packSizes) == 0 {
		return nil, entity.ErrEmptyOrder
	}

	combination := s.calculateOptimalCombination(req.Amount, packSizes)

	totalPacks := 0
	totalAmount := 0
	for size, quantity := range combination {
		totalPacks += quantity
		totalAmount += size * quantity
	}

	s.logger.Info("Optimal pack calculation completed - Total packs: %d, Total amount: %d", totalPacks, totalAmount)

	return &PackCalculationResponse{
		Amount:      req.Amount,
		PackSizes:   packSizes,
		Combination: combination,
		TotalPacks:  totalPacks,
		TotalAmount: totalAmount,
	}, nil
}

// calculateOptimalCombination finds the optimal pack combination following the rules:
// 1. Only whole packs can be sent
// 2. Send out the least amount of items to fulfill the order (minimize waste)
// 3. Send out as few packs as possible (minimize pack count)
// Rule #2 takes precedence over rule #3
func (s *PackService) calculateOptimalCombination(amount int, packSizes []int) map[int]int {
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))

	return s.findOptimalPacks(amount, sizes)
}

// findOptimalPacks implements an efficient pack optimization algorithm
func (s *PackService) findOptimalPacks(amount int, sizes []int) map[int]int {
	bestCombination := s.greedyApproach(amount, sizes)
	bestWaste, bestPackCount := s.calculateWasteAndPacks(amount, bestCombination)

	if amount <= 100000 || bestWaste > 0 {
		if zeroWasteSolution := s.searchZeroWasteSolution(amount, sizes); len(zeroWasteSolution) > 0 {
			waste, packCount := s.calculateWasteAndPacks(amount, zeroWasteSolution)
			if waste < bestWaste || (waste == bestWaste && packCount < bestPackCount) {
				return zeroWasteSolution
			}
		}
	}

	optimizedSolution := s.optimizeGreedySolution(amount, sizes, bestCombination)
	optimizedWaste, optimizedPackCount := s.calculateWasteAndPacks(amount, optimizedSolution)

	if optimizedWaste < bestWaste || (optimizedWaste == bestWaste && optimizedPackCount < bestPackCount) {
		return optimizedSolution
	}

	return bestCombination
}

// greedyApproach implements a greedy algorithm using largest pack sizes first
func (s *PackService) greedyApproach(amount int, sizes []int) map[int]int {
	combination := make(map[int]int)
	remaining := amount

	for _, size := range sizes {
		if remaining <= 0 {
			break
		}
		count := remaining / size
		if count > 0 {
			combination[size] = count
			remaining -= count * size
		}
	}

	if remaining > 0 && len(sizes) > 0 {
		smallestSize := sizes[len(sizes)-1]
		combination[smallestSize]++
	}

	return combination
}

// optimizeGreedySolution tries to improve the greedy solution
func (s *PackService) optimizeGreedySolution(amount int, sizes []int, greedySolution map[int]int) map[int]int {
	bestCombination := s.copyMap(greedySolution)
	bestWaste, bestPackCount := s.calculateWasteAndPacks(amount, bestCombination)

	for i := 0; i < len(sizes)-1; i++ {
		for j := i + 1; j < len(sizes); j++ {
			largerSize, smallerSize := sizes[i], sizes[j]

			if bestCombination[largerSize] > 0 {
				testCombination := s.copyMap(bestCombination)
				testCombination[largerSize]--
				if testCombination[largerSize] == 0 {
					delete(testCombination, largerSize)
				}

				needed := largerSize
				smallerCount := (needed + smallerSize - 1) / smallerSize
				testCombination[smallerSize] += smallerCount

				waste, packCount := s.calculateWasteAndPacks(amount, testCombination)
				if waste < bestWaste || (waste == bestWaste && packCount < bestPackCount) {
					bestWaste, bestPackCount = waste, packCount
					bestCombination = testCombination
				}
			}
		}
	}

	return bestCombination
}

// copyMap creates a deep copy of a map[int]int
func (s *PackService) copyMap(original map[int]int) map[int]int {
	copy := make(map[int]int)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// searchZeroWasteSolution searches for combinations that result in zero waste
func (s *PackService) searchZeroWasteSolution(amount int, sizes []int) map[int]int {
	if len(sizes) == 0 {
		return make(map[int]int)
	}

	if amount <= 50000 {
		return s.findExactSolution(amount, sizes)
	}

	return s.findLargeAmountSolution(amount, sizes)
}

// findExactSolution tries to find an exact solution for smaller amounts
func (s *PackService) findExactSolution(amount int, sizes []int) map[int]int {
	return s.recursiveExactSearch(amount, sizes, 0, make(map[int]int))
}

// findLargeAmountSolution uses a systematic approach for larger amounts
func (s *PackService) findLargeAmountSolution(amount int, sizes []int) map[int]int {
	largestSize := sizes[0]
	maxLargest := amount / largestSize

	for largestCount := maxLargest; largestCount >= maxLargest-20 && largestCount >= 0; largestCount-- {
		remaining := amount - (largestSize * largestCount)
		if remaining == 0 {
			return map[int]int{largestSize: largestCount}
		}
		if remaining > 0 {
			if solution := s.solveRemaining(remaining, sizes[1:]); len(solution) > 0 {
				solution[largestSize] = largestCount
				return solution
			}
		}
	}

	return make(map[int]int)
}

// recursiveExactSearch performs a recursive search for exact solutions
func (s *PackService) recursiveExactSearch(amount int, sizes []int, index int, current map[int]int) map[int]int {
	if amount == 0 {
		return s.copyMap(current)
	}
	if amount < 0 || index >= len(sizes) {
		return make(map[int]int)
	}

	size := sizes[index]
	maxCount := amount / size

	for count := 0; count <= maxCount; count++ {
		newCurrent := s.copyMap(current)
		if count > 0 {
			newCurrent[size] = count
		}

		if result := s.recursiveExactSearch(amount-size*count, sizes, index+1, newCurrent); len(result) > 0 {
			return result
		}
	}

	return make(map[int]int)
}

// solveRemaining tries to solve the remaining amount with given sizes
func (s *PackService) solveRemaining(amount int, sizes []int) map[int]int {
	if len(sizes) == 0 {
		return make(map[int]int)
	}
	if len(sizes) == 1 {
		if amount%sizes[0] == 0 {
			return map[int]int{sizes[0]: amount / sizes[0]}
		}
		return make(map[int]int)
	}

	return s.recursiveExactSearch(amount, sizes, 0, make(map[int]int))
}

// calculateWasteAndPacks calculates waste and pack count for a given combination
func (s *PackService) calculateWasteAndPacks(amount int, combination map[int]int) (int, int) {
	total := 0
	packCount := 0

	for size, count := range combination {
		total += size * count
		packCount += count
	}

	waste := 0
	if total > amount {
		waste = total - amount
	}

	return waste, packCount
}

// GetAllPacks returns all available packs
func (s *PackService) GetAllPacks(ctx context.Context) []entity.Pack {
	s.logger.Debug("Getting all packs")

	packs := s.packRepo.List(ctx)
	s.logger.Debug("Retrieved %d packs", len(packs))

	return packs
}

// CreatePack creates a new pack
func (s *PackService) CreatePack(ctx context.Context, pack *entity.Pack) error {
	exists, err := s.packRepo.ExistsBySize(ctx, pack.Size())
	if err != nil {
		s.logger.Error("Failed to check if pack size exists: %v", err)
		return err
	}

	if exists {
		s.logger.Warn("Attempted to create pack with duplicate size: %d", pack.Size())
		return entity.ErrDuplicatePackSize
	}

	return s.packRepo.Create(ctx, pack)
}

// UpdatePack updates an existing pack
func (s *PackService) UpdatePack(ctx context.Context, pack *entity.Pack) error {
	currentPack, err := s.packRepo.Get(ctx, pack.ID())
	if err != nil {
		s.logger.Error("Failed to get current pack for update: %v", err)
		return err
	}

	if pack.Size() != currentPack.Size() {
		exists, err := s.packRepo.ExistsBySize(ctx, pack.Size())
		if err != nil {
			s.logger.Error("Failed to check if pack size exists during update: %v", err)
			return err
		}

		if exists {
			s.logger.Warn("Attempted to update pack to duplicate size: %d", pack.Size())
			return entity.ErrDuplicatePackSize
		}
	}

	return s.packRepo.Update(ctx, pack)
}

// DeletePack deletes a pack
func (s *PackService) DeletePack(ctx context.Context, pack *entity.Pack) error {
	s.logger.Info("Deleting pack with ID: %s, size: %d", pack.ID(), pack.Size())

	err := s.packRepo.Delete(ctx, pack)
	if err != nil {
		s.logger.Error("Failed to delete pack %s: %v", pack.ID(), err)
		return err
	}

	s.logger.Info("Pack deleted successfully with ID: %s", pack.ID())
	return nil
}

// GetPackByID retrieves a pack by its ID
func (s *PackService) GetPackByID(ctx context.Context, id string) (*entity.Pack, error) {
	s.logger.Debug("Getting pack by ID: %s", id)

	packs := s.packRepo.List(ctx)
	for _, pack := range packs {
		if pack.ID().String() == id {
			s.logger.Debug("Pack found with ID: %s", id)
			return &pack, nil
		}
	}

	s.logger.Warn("Pack not found with ID: %s", id)
	return nil, entity.ErrOrderNotFound
}
