package system

import (
	"golang.org/x/sys/windows/registry"
)

// `RegistryTable` represents a database of information gathered from the clients Windows registry.
type RegistryTable struct {
	HWID            string `json:"hardware_id"`
	Version         string `json:"windows_version"`
	CurrentUserName string `json:"current_user_name"`
	ProductName     string `json:"product_name"`
	ProductId       string `json:"product_id"`
	InstallDate     string `json:"install_date"`
	RegisteredOwner string `json:"registered_owner"`
	RegisteredOrg   string `json:"registered_organization"`
}

// `AddToStartup` adds the passed file path to the Windows startup registry key, so that
// the file is executed automatically when the machine initially starts up.
func AddFileToStartup(path, appName string) error {
	// Open the CurrentVersion/Run registry key to add a new application to.
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	// Set the string value of new application (appName) to the passed executable path (path).
	if err := key.SetStringValue(appName, path); err != nil {
		return err
	}

	return nil
}

// `GetHardwareID` returns the hardware ID of the machine by reading the "MachineGuid" value
// from the Cryptography registry key.
func GetHardwareID() string {
	var hwid = DEFAULT_REG_VAL

	// Open the Cryptography registry key and get the HWID value from it.
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return hwid
	}
	defer key.Close()

	// Read the value of the "MachineGuid" key to get the hardware ID.
	hwid, _, err = key.GetStringValue("MachineGuid")
	if err != nil {
		return hwid
	}

	return hwid
}

// `GetRegistryInformation` returns a RegistryTable struct containing information about the
// Windows system by reading values from the registry, including the hardware ID, Windows
// version, product name and ID, installation date, registered owner, and registered organization.
func GetRegistryInformation() RegistryTable {
	// Set the default RegistryTable values using the DEFAULT_REG_VAL constant.
	var registryTable = RegistryTable{
		HWID:            DEFAULT_REG_VAL,
		Version:         DEFAULT_REG_VAL,
		CurrentUserName: DEFAULT_REG_VAL,
		ProductName:     DEFAULT_REG_VAL,
		InstallDate:     DEFAULT_REG_VAL,
		RegisteredOwner: DEFAULT_REG_VAL,
		RegisteredOrg:   DEFAULT_REG_VAL,
	}

	registryTable.HWID = GetHardwareID()

	// Open the CurrentVersion registry key and get its values.
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ)
	if err == nil {
		// Read the value of the "CurrentVersion" key to get the Windows version.
		version, _, err := key.GetStringValue("CurrentVersion")
		if err == nil {
			registryTable.Version = version
		}

		// Read the value of the "ProductName" key to get the Windows product name.
		productName, _, err := key.GetStringValue("ProductName")
		if err == nil {
			registryTable.ProductName = productName
		}

		// Read the value of the "ProductId" key to get the Windows product ID.
		product_id, _, err := key.GetStringValue("ProductId")
		if err == nil {
			registryTable.ProductId = product_id
		}

		// Read the value of the "InstallDate" key to get the Windows installation date.
		installDate, _, err := key.GetStringValue("InstallDate")
		if err == nil {
			registryTable.InstallDate = installDate
		}

		// Read the value of the "RegisteredOwner" key to get the registered owner of the Windows product.
		registeredOwner, _, err := key.GetStringValue("RegisteredOwner")
		if err == nil {
			registryTable.RegisteredOwner = registeredOwner
		}

		// Read the value of the "RegisteredOrganization" key to get the registered organization of the Windows product.
		registeredOrg, _, err := key.GetStringValue("RegisteredOrganization")
		if err == nil {
			registryTable.RegisteredOrg = registeredOrg
		}

		key.Close()
	}

	return registryTable
}

// Default value to use in the RegistryTable when the queried value isn't valid.
const DEFAULT_REG_VAL = "Unknown"
