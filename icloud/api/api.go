package api

import (
	"fmt"
	"strings"
)

type dict map[string]interface{}

// Codes returned by API
const (
	CodeWrongVerification  = 21669
	CodeWrongVerification2 = -21669
	CodeNotFound           = 404
)

// Device describes a user device like iPhone, iPad and so on
type Device struct {
	DeviceType  string `json:"deviceType"`
	AreaCode    string `json:"areaCode"`
	PhoneNumber string `json:"phoneNumber"`
	DeviceID    string `json:"deviceId"`
}

func (d *Device) Dict() dict {
	return dict{
		"deviceType":  d.DeviceType,
		"areaCode":    d.AreaCode,
		"phoneNumber": d.PhoneNumber,
		"deviceId":    d.DeviceID,
	}
}

// DeviceResponse ...
type DeviceResponse struct {
	Devices []Device `json:"devices"`
}

// DsInfo ...
type DsInfo struct {
	ADsID          string        `json:"aDsID"`
	AppleID        string        `json:"appleId"`
	AppleIDAlias   string        `json:"appleIdAlias"`
	AppleIDAliases []interface{} `json:"appleIdAliases"`
	AppleIDEntries []struct {
		IsPrimary bool   `json:"isPrimary"`
		Type      string `json:"type"`
		Value     string `json:"value"`
	} `json:"appleIdEntries"`
	BeneficiaryInfo struct {
		IsBeneficiary bool `json:"isBeneficiary"`
	} `json:"beneficiaryInfo"`
	BrMigrated                      bool   `json:"brMigrated"`
	BrZoneConsolidated              bool   `json:"brZoneConsolidated"`
	CountryCode                     string `json:"countryCode"`
	Dsid                            string `json:"dsid"`
	FamilyEligible                  bool   `json:"familyEligible"`
	FirstName                       string `json:"firstName"`
	FullName                        string `json:"fullName"`
	GilliganEnabled                 bool   `json:"gilligan-enabled"`
	GilliganInvited                 bool   `json:"gilligan-invited"`
	HasICloudQualifyingDevice       bool   `json:"hasICloudQualifyingDevice"`
	HasPaymentInfo                  bool   `json:"hasPaymentInfo"`
	HsaEnabled                      bool   `json:"hsaEnabled"`
	HsaVersion                      int    `json:"hsaVersion"`
	ICDPEnabled                     bool   `json:"iCDPEnabled"`
	ICloudAppleIDAlias              string `json:"iCloudAppleIdAlias"`
	IroncadeMigrated                bool   `json:"ironcadeMigrated"`
	IsCustomDomainsFeatureAvailable bool   `json:"isCustomDomainsFeatureAvailable"`
	IsHideMyEmailFeatureAvailable   bool   `json:"isHideMyEmailFeatureAvailable"`
	IsManagedAppleID                bool   `json:"isManagedAppleID"`
	IsPaidDeveloper                 bool   `json:"isPaidDeveloper"`
	LanguageCode                    string `json:"languageCode"`
	LastName                        string `json:"lastName"`
	Locale                          string `json:"locale"`
	Locked                          bool   `json:"locked"`
	NotesMigrated                   bool   `json:"notesMigrated"`
	NotificationID                  string `json:"notificationId"`
	PcsDeleted                      bool   `json:"pcsDeleted"`
	PrimaryEmail                    string `json:"primaryEmail"`
	PrimaryEmailVerified            bool   `json:"primaryEmailVerified"`
	StatusCode                      int    `json:"statusCode"`
	TantorMigrated                  bool   `json:"tantorMigrated"`
}

// Webservices ...
type Webservices struct {
	Account struct {
		ICloudEnv struct {
			ShortID   string `json:"shortId"`
			VipSuffix string `json:"vipSuffix"`
		} `json:"iCloudEnv"`
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"account"`
	Ckdatabasews struct {
		PcsRequired bool   `json:"pcsRequired"`
		Status      string `json:"status"`
		URL         string `json:"url"`
	} `json:"ckdatabasews"`
	Ckdeviceservice struct {
		URL string `json:"url"`
	} `json:"ckdeviceservice"`
	Cksharews struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"cksharews"`
	Contacts struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"contacts"`
	Docws struct {
		PcsRequired bool   `json:"pcsRequired"`
		Status      string `json:"status"`
		URL         string `json:"url"`
	} `json:"docws"`
	Drivews struct {
		PcsRequired bool   `json:"pcsRequired"`
		Status      string `json:"status"`
		URL         string `json:"url"`
	} `json:"drivews"`
	Geows struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"geows"`
	Iwmb struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"iwmb"`
	Iworkexportws struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"iworkexportws"`
	Iworkthumbnailws struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"iworkthumbnailws"`
	Keyvalue struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"keyvalue"`
	Premiummailsettings struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"premiummailsettings"`
	Push struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"push"`
	Schoolwork struct {
	} `json:"schoolwork"`
	Settings struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"settings"`
	Ubiquity struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"ubiquity"`
	Uploadimagews struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	} `json:"uploadimagews"`
}

func (w *Webservices) URL(service string) (url string, err error) {
	shortName := strings.TrimSuffix(strings.ToLower(service), "ws")
	switch shortName {
	case "account":
		url = w.Account.URL
	case "ckdatabase":
		url = w.Ckdatabasews.URL
	case "ckdeviceservice":
		url = w.Ckdeviceservice.URL
	case "ckshare":
		url = w.Cksharews.URL
	case "contacts":
		url = w.Contacts.URL
	case "doc":
		url = w.Docws.URL
	case "drive":
		url = w.Drivews.URL
	case "geo":
		url = w.Geows.URL
	case "iwmb":
		url = w.Iwmb.URL
	case "iworkexport":
		url = w.Iworkexportws.URL
	case "iworkthumbnail":
		url = w.Iworkthumbnailws.URL
	case "keyvalue":
		url = w.Keyvalue.URL
	case "premiummailsettings":
		url = w.Premiummailsettings.URL
	case "push":
		url = w.Push.URL
	case "settings":
		url = w.Settings.URL
	case "ubiquity":
		url = w.Ubiquity.URL
	case "uploadimage":
		url = w.Uploadimagews.URL
	}
	if url == "" {
		err = fmt.Errorf("service %q does not have an URL", service)
	}
	return
}

// Apps ...
type Apps struct {
	Contacts struct {
	} `json:"contacts"`
	Find struct {
		CanLaunchWithOneFactor bool `json:"canLaunchWithOneFactor"`
	} `json:"find"`
	Iclouddrive struct {
	} `json:"iclouddrive"`
	Keynote struct {
		IsQualifiedForBeta bool `json:"isQualifiedForBeta"`
	} `json:"keynote"`
	Newspublisher struct {
		IsHidden bool `json:"isHidden"`
	} `json:"newspublisher"`
	Notes3 struct {
	} `json:"notes3"`
	Numbers struct {
		IsQualifiedForBeta bool `json:"isQualifiedForBeta"`
	} `json:"numbers"`
	Pages struct {
		IsQualifiedForBeta bool `json:"isQualifiedForBeta"`
	} `json:"pages"`
	Settings struct {
		CanLaunchWithOneFactor bool `json:"canLaunchWithOneFactor"`
	} `json:"settings"`
}

func (a *Apps) AllowsOneFactor(service string) (can bool, err error) {
	switch strings.ToLower(service) {
	case "find":
		can = a.Find.CanLaunchWithOneFactor
	case "settings":
		can = a.Settings.CanLaunchWithOneFactor
	default:
		err = fmt.Errorf("service %q does not allow 1-factor", service)
	}
	return
}

// ConfigBag ...
type ConfigBag struct {
	AccountCreateEnabled string `json:"accountCreateEnabled"`
	Urls                 struct {
		AccountAuthorizeUI  string `json:"accountAuthorizeUI"`
		AccountCreate       string `json:"accountCreate"`
		AccountCreateUI     string `json:"accountCreateUI"`
		AccountLogin        string `json:"accountLogin"`
		AccountLoginUI      string `json:"accountLoginUI"`
		AccountRepairUI     string `json:"accountRepairUI"`
		DownloadICloudTerms string `json:"downloadICloudTerms"`
		GetICloudTerms      string `json:"getICloudTerms"`
		RepairDone          string `json:"repairDone"`
		VettingURLForEmail  string `json:"vettingUrlForEmail"`
		VettingURLForPhone  string `json:"vettingUrlForPhone"`
	} `json:"urls"`
}

// ICloudInfo ...
type ICloudInfo struct {
	SafariBookmarksHasMigratedToCloudKit bool `json:"SafariBookmarksHasMigratedToCloudKit"`
}

// RequestInfo ...
type RequestInfo struct {
	Country  string `json:"country"`
	Region   string `json:"region"`
	TimeZone string `json:"timeZone"`
}

// StateResponse ...
type StateResponse struct {
	Apps                         Apps        `json:"apps"`
	AppsOrder                    []string    `json:"appsOrder"`
	ConfigBag                    ConfigBag   `json:"configBag"`
	DsInfo                       DsInfo      `json:"dsInfo"`
	HasMinimumDeviceForPhotosWeb bool        `json:"hasMinimumDeviceForPhotosWeb"`
	HsaChallengeRequired         bool        `json:"hsaChallengeRequired"`
	HsaTrustedBrowser            bool        `json:"hsaTrustedBrowser"`
	ICDPEnabled                  bool        `json:"iCDPEnabled"`
	ICLoudInfo                   ICloudInfo  `json:"iCloudInfo"`
	IsExtendedLogin              bool        `json:"isExtendedLogin"`
	IsRepairNeeded               bool        `json:"isRepairNeeded"`
	PcsDeleted                   bool        `json:"pcsDeleted"`
	PcsEnabled                   bool        `json:"pcsEnabled"`
	PcsServiceIdentitiesIncluded bool        `json:"pcsServiceIdentitiesIncluded"`
	RequestInfo                  RequestInfo `json:"requestInfo"`
	TermsUpdateNeeded            bool        `json:"termsUpdateNeeded"`
	Version                      int         `json:"version"`
	Webservices                  Webservices `json:"webservices"`
}

// ErrorResponse ...
type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
	Reason       string `json:"reason"`
	ErrorReason  string `json:"errorReason"`
	Error        string `json:"error"`
	Code         int    `json:"errorCode"`
	ServerCode   int    `json:"serverErrorCode"`
}

// SuccessResponse ...
type SuccessResponse struct {
	Success bool `json:"success"`
}
