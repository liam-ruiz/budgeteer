package util

import (
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)


func NumericToString(n pgtype.Numeric) string {
	if !n.Valid {
        return "0.00"
    }

    // Convert pgtype.Numeric to float64
    f, err := n.Float64Value()
    if err != nil {
        return "Error"
    }

    // Standard formatting (or use the message printer above for commas)
    return fmt.Sprintf("%.2f", f.Float64)
}

func Float64ToNumeric(f float64) pgtype.Numeric {
	// convert to bigint
	bigInt := big.NewInt(int64(f * 100))
	return pgtype.Numeric{
		Valid: true,
		Int: bigInt,
	}
}
	