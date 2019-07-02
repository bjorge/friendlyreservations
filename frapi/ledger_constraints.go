package frapi

const updateBalanceConstraintsGQL = `

type UpdateBalanceConstraints {
	amountMin: Int!
	amountMax: Int!
	descriptionMin: Int!
	descriptionMax: Int!
}
`

// UpdateBalanceConstraints provides the retriever for contraint methods
type UpdateBalanceConstraints struct {
}

// UpdateBalanceConstraints returns input constraints for updating Balance
func (r *PropertyResolver) UpdateBalanceConstraints() (*UpdateBalanceConstraints, error) {

	return &UpdateBalanceConstraints{}, nil
}

// AmountMin returns the minimum amount
func (r *UpdateBalanceConstraints) AmountMin() int32 { return 1 }

// AmountMax returns the maximum amount
func (r *UpdateBalanceConstraints) AmountMax() int32 { return 100000 }

// DescriptionMin returns the minimum description length
func (r *UpdateBalanceConstraints) DescriptionMin() int32 { return 3 }

// DescriptionMax returns the maximum description length
func (r *UpdateBalanceConstraints) DescriptionMax() int32 { return 35 }
