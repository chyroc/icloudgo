package internal

import (
	"encoding/json"
	"fmt"
)

func (r *PhotoAlbum) PhotosIter(startOffset int) PhotosIterNext {
	if r.Direction == "DESCENDING" {
		startOffset = r.Size() - 1 - startOffset
	}
	return newPhotosIterNext(r, startOffset)
}

func (r *PhotoAlbum) GetPhotosByOffset(offset, limit int) ([]*PhotoAsset, error) {
	var assets []*PhotoAsset

	text, err := r.service.icloud.request(&rawReq{
		Method:  "POST",
		URL:     fmt.Sprintf("%s/records/query", r.service.serviceEndpoint),
		Querys:  r.service.querys,
		Headers: r.service.icloud.getCommonHeaders(map[string]string{}),
		Body:    r.listQueryGenerate(offset, limit, r.ListType, r.Direction, r.QueryFilter),
	})
	if err != nil {
		return nil, fmt.Errorf("get album photos failed, err: %w", err)
	}
	res := new(getPhotosResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return nil, fmt.Errorf("get album photos unmarshal failed, err: %w", err)
	}

	var masterRecords []*photoRecord
	assetRecords := map[string]*photoRecord{}
	for _, record := range res.Records {
		if record.RecordType == "CPLAsset" {
			masterID := record.Fields.MasterRef.Value.RecordName
			assetRecords[masterID] = record
		} else if record.RecordType == "CPLMaster" {
			masterRecords = append(masterRecords, record)
		}
	}

	for _, masterRecord := range masterRecords {
		assets = append(assets,
			r.service.newPhotoAsset(masterRecord, assetRecords[masterRecord.RecordName]),
		)
	}

	return assets, nil
}

func (r *PhotoAlbum) GetPhotosByCount(count int) ([]*PhotoAsset, error) {
	offset := 0
	if r.Direction == "DESCENDING" {
		offset = r.Size() - 1
	}

	var assets []*PhotoAsset
	for {
		tmp, err := r.GetPhotosByOffset(offset, 200)
		if err != nil {
			return nil, err
		}
		if len(tmp) == 0 {
			break
		}
		for _, v := range tmp {
			assets = append(assets, v)
			if len(assets) >= count {
				return assets, nil
			}
		}
		offset = r.calOffset(offset, len(tmp))
	}

	return assets, nil
}

func (r *PhotoAlbum) WalkPhotos(offset int, f func(offset int, assets []*PhotoAsset) error) error {
	size := r.Size()
	if r.Direction == "DESCENDING" {
		offset = size - 1 - offset
	}
	for {
		tmp, err := r.GetPhotosByOffset(offset, 200)
		if err != nil {
			return err
		}
		fmt.Printf("[icloudgo] [walk_photo] name: %s, offset: %d, size=%d, got=%d, desc=%v\n", r.Name, offset, size, len(tmp), r.Direction == "DESCENDING")
		if len(tmp) == 0 {
			break
		}
		offset = r.calOffset(offset, len(tmp))

		if err := f(offset, tmp); err != nil {
			return err
		}
	}
	return nil
}

func (r *PhotoAlbum) calOffset(offset, lastAssetLen int) int {
	if r.Direction == "DESCENDING" {
		offset = offset - lastAssetLen
	} else {
		offset = offset + lastAssetLen
	}
	return offset
}

func (r *PhotoAlbum) listQueryGenerate(offset, limit int, listType string, direction string, queryFilter []*folderMetaDataQueryFilter) any {
	res := map[string]any{
		"query": map[string]any{
			"filterBy": append([]*folderMetaDataQueryFilter{
				{
					FieldName:  "startRank",
					FieldValue: &folderTypeValue{Type: "INT64", Value: offset},
					Comparator: "EQUALS",
				},
				{
					FieldName:  "direction",
					FieldValue: &folderTypeValue{Type: "STRING", Value: direction},
					Comparator: "EQUALS",
				},
			}, queryFilter...),
			"recordType": listType,
		},
		"resultsLimit": limit,
		"desiredKeys": []string{
			"resJPEGFullWidth",
			"resJPEGFullHeight",
			"resJPEGFullFileType",
			"resJPEGFullFingerprint",
			"resJPEGFullRes",
			"resJPEGLargeWidth",
			"resJPEGLargeHeight",
			"resJPEGLargeFileType",
			"resJPEGLargeFingerprint",
			"resJPEGLargeRes",
			"resJPEGMedWidth",
			"resJPEGMedHeight",
			"resJPEGMedFileType",
			"resJPEGMedFingerprint",
			"resJPEGMedRes",
			"resJPEGThumbWidth",
			"resJPEGThumbHeight",
			"resJPEGThumbFileType",
			"resJPEGThumbFingerprint",
			"resJPEGThumbRes",
			"resVidFullWidth",
			"resVidFullHeight",
			"resVidFullFileType",
			"resVidFullFingerprint",
			"resVidFullRes",
			"resVidMedWidth",
			"resVidMedHeight",
			"resVidMedFileType",
			"resVidMedFingerprint",
			"resVidMedRes",
			"resVidSmallWidth",
			"resVidSmallHeight",
			"resVidSmallFileType",
			"resVidSmallFingerprint",
			"resVidSmallRes",
			"resSidecarWidth",
			"resSidecarHeight",
			"resSidecarFileType",
			"resSidecarFingerprint",
			"resSidecarRes",
			"itemType",
			"dataClassType",
			"filenameEnc",
			"originalOrientation",
			"resOriginalWidth",
			"resOriginalHeight",
			"resOriginalFileType",
			"resOriginalFingerprint",
			"resOriginalRes",
			"resOriginalAltWidth",
			"resOriginalAltHeight",
			"resOriginalAltFileType",
			"resOriginalAltFingerprint",
			"resOriginalAltRes",
			"resOriginalVidComplWidth",
			"resOriginalVidComplHeight",
			"resOriginalVidComplFileType",
			"resOriginalVidComplFingerprint",
			"resOriginalVidComplRes",
			"isDeleted",
			"isExpunged",
			"dateExpunged",
			"remappedRef",
			"recordName",
			"recordType",
			"recordChangeTag",
			"masterRef",
			"adjustmentRenderType",
			"assetDate",
			"addedDate",
			"isFavorite",
			"isHidden",
			"orientation",
			"duration",
			"assetSubtype",
			"assetSubtypeV2",
			"assetHDRType",
			"burstFlags",
			"burstFlagsExt",
			"burstId",
			"captionEnc",
			"locationEnc",
			"locationV2Enc",
			"locationLatitude",
			"locationLongitude",
			"adjustmentType",
			"timeZoneOffset",
			"vidComplDurValue",
			"vidComplDurScale",
			"vidComplDispValue",
			"vidComplDispScale",
			"vidComplVisibilityState",
			"customRenderedValue",
			"containerId",
			"itemId",
			"position",
			"isKeyAsset",
		},
		"zoneID": map[string]any{"zoneName": "PrimarySync"},
	}

	return res
}

type getPhotosResp struct {
	Records            []*photoRecord `json:"records"`
	ContinuationMarker string         `json:"continuationMarker"`
	SyncToken          string         `json:"syncToken"`
}

type photoRecord struct {
	RecordName string `json:"recordName"`
	RecordType string `json:"recordType"`
	Fields     struct {
		ItemType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"itemType,omitempty"`
		ResJPEGThumbFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGThumbFingerprint,omitempty"`
		FilenameEnc struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"filenameEnc,omitempty"`
		ResJPEGMedRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resJPEGMedRes,omitempty"`
		OriginalOrientation struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"originalOrientation,omitempty"`
		ResJPEGMedHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGMedHeight,omitempty"`
		ResOriginalRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resOriginalRes,omitempty"`
		ResJPEGMedFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGMedFileType,omitempty"`
		ResJPEGThumbHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGThumbHeight,omitempty"`
		ResJPEGThumbWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGThumbWidth,omitempty"`
		ResOriginalWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalWidth,omitempty"`
		ResJPEGThumbFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGThumbFileType,omitempty"`
		DataClassType struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"dataClassType,omitempty"`
		ResOriginalFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalFingerprint,omitempty"`
		ResJPEGMedWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGMedWidth,omitempty"`
		ResJPEGThumbRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resJPEGThumbRes,omitempty"`
		ResOriginalFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalFileType,omitempty"`
		ResOriginalHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalHeight,omitempty"`
		ResJPEGMedFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resJPEGMedFingerprint,omitempty"`
		ResVidSmallHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidSmallHeight,omitempty"`
		ResOriginalVidComplFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalVidComplFileType,omitempty"`
		ResOriginalVidComplWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalVidComplWidth,omitempty"`
		ResVidMedFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidMedFileType,omitempty"`
		ResVidMedRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resVidMedRes,omitempty"`
		ResVidSmallFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidSmallFingerprint,omitempty"`
		ResVidMedWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidMedWidth,omitempty"`
		ResOriginalVidComplFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalVidComplFingerprint,omitempty"`
		ResVidSmallFileType struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidSmallFileType,omitempty"`
		ResVidSmallRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resVidSmallRes,omitempty"`
		ResOriginalVidComplRes struct {
			Value struct {
				FileChecksum      string `json:"fileChecksum"`
				Size              int    `json:"size"`
				WrappingKey       string `json:"wrappingKey"`
				ReferenceChecksum string `json:"referenceChecksum"`
				DownloadURL       string `json:"downloadURL"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"resOriginalVidComplRes,omitempty"`
		ResVidMedFingerprint struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidMedFingerprint,omitempty"`
		ResVidMedHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidMedHeight,omitempty"`
		ResOriginalVidComplHeight struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resOriginalVidComplHeight,omitempty"`
		ResVidSmallWidth struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"resVidSmallWidth,omitempty"`
		AssetDate struct {
			Value int64  `json:"value"`
			Type  string `json:"type"`
		} `json:"assetDate,omitempty"`
		Orientation struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"orientation,omitempty"`
		AddedDate struct {
			Value int64  `json:"value"`
			Type  string `json:"type"`
		} `json:"addedDate,omitempty"`
		AssetSubtypeV2 struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"assetSubtypeV2,omitempty"`
		AssetHDRType struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"assetHDRType,omitempty"`
		TimeZoneOffset struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"timeZoneOffset,omitempty"`
		MasterRef struct {
			Value struct {
				RecordName string `json:"recordName"`
				Action     string `json:"action"`
				ZoneID     struct {
					ZoneName        string `json:"zoneName"`
					OwnerRecordName string `json:"ownerRecordName"`
					ZoneType        string `json:"zoneType"`
				} `json:"zoneID"`
			} `json:"value"`
			Type string `json:"type"`
		} `json:"masterRef,omitempty"`
		AdjustmentRenderType struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"adjustmentRenderType,omitempty"`
		VidComplDispScale struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"vidComplDispScale,omitempty"`
		IsHidden struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"isHidden,omitempty"`
		Duration struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"duration,omitempty"`
		BurstFlags struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"burstFlags,omitempty"`
		AssetSubtype struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"assetSubtype,omitempty"`
		VidComplDurScale struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"vidComplDurScale,omitempty"`
		VidComplDurValue struct {
			Value int64  `json:"value"`
			Type  string `json:"type"`
		} `json:"vidComplDurValue,omitempty"`
		VidComplVisibilityState struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"vidComplVisibilityState,omitempty"`
		CustomRenderedValue struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"customRenderedValue,omitempty"`
		IsFavorite struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"isFavorite,omitempty"`
		VidComplDispValue struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"vidComplDispValue,omitempty"`
		LocationEnc struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"locationEnc,omitempty"`
	} `json:"fields"`
	PluginFields    struct{} `json:"pluginFields"`
	RecordChangeTag string   `json:"recordChangeTag"`
	Created         struct {
		Timestamp      int64  `json:"timestamp"`
		UserRecordName string `json:"userRecordName"`
		DeviceID       string `json:"deviceID"`
	} `json:"created"`
	Modified struct {
		Timestamp      int64  `json:"timestamp"`
		UserRecordName string `json:"userRecordName"`
		DeviceID       string `json:"deviceID"`
	} `json:"modified"`
	Deleted bool `json:"deleted"`
	ZoneID  struct {
		ZoneName        string `json:"zoneName"`
		OwnerRecordName string `json:"ownerRecordName"`
		ZoneType        string `json:"zoneType"`
	} `json:"zoneID"`
}
