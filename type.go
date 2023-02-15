package icloudgo

type User struct {
	AccountName string `json:"accountName"`
	Password    string `json:"password"`
}

type SessionData struct {
	SessionToken   string `json:"session_token"`
	Scnt           string `json:"scnt"`
	SessionID      string `json:"session_id"`
	AccountCountry string `json:"account_country"`
	TrustToken     string `json:"trust_token"`
}

type ValidateData struct {
	DsInfo                       *ValidateDataDsInfo    `json:"dsInfo"`
	HasMinimumDeviceForPhotosWeb bool                   `json:"hasMinimumDeviceForPhotosWeb"`
	ICDPEnabled                  bool                   `json:"iCDPEnabled"`
	Webservices                  map[string]*webService `json:"webservices"`
	PcsEnabled                   bool                   `json:"pcsEnabled"`
	TermsUpdateNeeded            bool                   `json:"termsUpdateNeeded"`
	ConfigBag                    struct {
		Urls struct {
			AccountCreateUI     string `json:"accountCreateUI"`
			AccountLoginUI      string `json:"accountLoginUI"`
			AccountLogin        string `json:"accountLogin"`
			AccountRepairUI     string `json:"accountRepairUI"`
			DownloadICloudTerms string `json:"downloadICloudTerms"`
			RepairDone          string `json:"repairDone"`
			AccountAuthorizeUI  string `json:"accountAuthorizeUI"`
			VettingUrlForEmail  string `json:"vettingUrlForEmail"`
			AccountCreate       string `json:"accountCreate"`
			GetICloudTerms      string `json:"getICloudTerms"`
			VettingUrlForPhone  string `json:"vettingUrlForPhone"`
		} `json:"urls"`
		AccountCreateEnabled bool `json:"accountCreateEnabled"`
	} `json:"configBag"`
	HsaTrustedBrowser            bool     `json:"hsaTrustedBrowser"`
	AppsOrder                    []string `json:"appsOrder"`
	Version                      int      `json:"version"`
	IsExtendedLogin              bool     `json:"isExtendedLogin"`
	PcsServiceIdentitiesIncluded bool     `json:"pcsServiceIdentitiesIncluded"`
	IsRepairNeeded               bool     `json:"isRepairNeeded"`
	HsaChallengeRequired         bool     `json:"hsaChallengeRequired"`
	RequestInfo                  struct {
		Country  string `json:"country"`
		TimeZone string `json:"timeZone"`
		Region   string `json:"region"`
	} `json:"requestInfo"`
	PcsDeleted bool `json:"pcsDeleted"`
	ICloudInfo struct {
		SafariBookmarksHasMigratedToCloudKit bool `json:"SafariBookmarksHasMigratedToCloudKit"`
	} `json:"iCloudInfo"`
	Apps map[string]*ValidateDataApp `json:"apps"`
}

type ValidateDataDsInfo struct {
	HsaVersion                         int           `json:"hsaVersion"`
	LastName                           string        `json:"lastName"`
	ICDPEnabled                        bool          `json:"iCDPEnabled"`
	TantorMigrated                     bool          `json:"tantorMigrated"`
	Dsid                               string        `json:"dsid"`
	HsaEnabled                         bool          `json:"hsaEnabled"`
	IsHideMyEmailSubscriptionActive    bool          `json:"isHideMyEmailSubscriptionActive"`
	IroncadeMigrated                   bool          `json:"ironcadeMigrated"`
	Locale                             string        `json:"locale"`
	BrZoneConsolidated                 bool          `json:"brZoneConsolidated"`
	ICDRSCapableDeviceList             string        `json:"ICDRSCapableDeviceList"`
	IsManagedAppleID                   bool          `json:"isManagedAppleID"`
	IsCustomDomainsFeatureAvailable    bool          `json:"isCustomDomainsFeatureAvailable"`
	IsHideMyEmailFeatureAvailable      bool          `json:"isHideMyEmailFeatureAvailable"`
	ContinueOnDeviceEligibleDeviceInfo []string      `json:"ContinueOnDeviceEligibleDeviceInfo"`
	Gilligvited                        bool          `json:"gilligvited"`
	AppleIdAliases                     []interface{} `json:"appleIdAliases"`
	UbiquityEOLEnabled                 bool          `json:"ubiquityEOLEnabled"`
	IsPaidDeveloper                    bool          `json:"isPaidDeveloper"`
	CountryCode                        string        `json:"countryCode"`
	NotificationId                     string        `json:"notificationId"`
	PrimaryEmailVerified               bool          `json:"primaryEmailVerified"`
	ADsID                              string        `json:"aDsID"`
	Locked                             bool          `json:"locked"`
	ICDRSCapableDeviceCount            int           `json:"ICDRSCapableDeviceCount"`
	HasICloudQualifyingDevice          bool          `json:"hasICloudQualifyingDevice"`
	PrimaryEmail                       string        `json:"primaryEmail"`
	AppleIdEntries                     []struct {
		IsPrimary bool   `json:"isPrimary"`
		Type      string `json:"type"`
		Value     string `json:"value"`
	} `json:"appleIdEntries"`
	GilliganEnabled    bool   `json:"gilligan-enabled"`
	IsWebAccessAllowed bool   `json:"isWebAccessAllowed"`
	FullName           string `json:"fullName"`
	MailFlags          struct {
		IsThreadingAvailable           bool `json:"isThreadingAvailable"`
		IsSearchV2Provisioned          bool `json:"isSearchV2Provisioned"`
		SCKMail                        bool `json:"sCKMail"`
		IsMppSupportedInCurrentCountry bool `json:"isMppSupportedInCurrentCountry"`
	} `json:"mailFlags"`
	LanguageCode         string `json:"languageCode"`
	AppleId              string `json:"appleId"`
	HasUnreleasedOS      bool   `json:"hasUnreleasedOS"`
	AnalyticsOptInStatus bool   `json:"analyticsOptInStatus"`
	FirstName            string `json:"firstName"`
	ICloudAppleIdAlias   string `json:"iCloudAppleIdAlias"`
	NotesMigrated        bool   `json:"notesMigrated"`
	BeneficiaryInfo      struct {
		IsBeneficiary bool `json:"isBeneficiary"`
	} `json:"beneficiaryInfo"`
	HasPaymentInfo bool   `json:"hasPaymentInfo"`
	PcsDelet       bool   `json:"pcsDelet"`
	AppleIdAlias   string `json:"appleIdAlias"`
	BrMigrated     bool   `json:"brMigrated"`
	StatusCode     int    `json:"statusCode"`
	FamilyEligible bool   `json:"familyEligible"`
}

type ValidateDataApp struct {
	CanLaunchWithOneFactor bool `json:"canLaunchWithOneFactor"` // Find
	IsQualifiedForBeta     bool `json:"isQualifiedForBeta"`     // Numbers
}

type webService struct {
	PcsRequired bool   `json:"pcsRequired"`
	URL         string `json:"url"`
	UploadURL   string `json:"uploadUrl"`
	Status      string `json:"status"`
}
