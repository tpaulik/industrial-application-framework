package parameters

import (
	common_types "github.com/nokia/industrial-application-framework/application-lib/pkg/types"
	"reflect"
)

func NetworkParametersChanged(instance common_types.OperatorCr) bool {

	// Comparing existing application parameters with new values in case of parameters whose value change is supported
	return !reflect.DeepEqual(instance.GetStatus().GetPrevSpec().GetPrivateNetworkAccess(), instance.GetSpec().GetPrivateNetworkAccess())
}
