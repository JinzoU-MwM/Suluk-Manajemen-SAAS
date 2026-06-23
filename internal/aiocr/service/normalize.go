package service

func validateExtractedData(data any, docType string) []ValidationError {
	var errors []ValidationError

	extracted, ok := data.(ExtractedFields)
	if !ok {
		return nil
	}

	switch docType {
	case "ktp":
		if extracted.NIK == "" {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK tidak terdeteksi"})
		} else if len(extracted.NIK) != 16 {
			errors = append(errors, ValidationError{Field: "nik", Message: "NIK harus 16 digit", Value: extracted.NIK})
		}
		if extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "paspor":
		if extracted.NoPaspor == "" {
			errors = append(errors, ValidationError{Field: "no_paspor", Message: "Nomor paspor tidak terdeteksi"})
		}
		if extracted.NamaPaspor == "" && extracted.Nama == "" {
			errors = append(errors, ValidationError{Field: "nama", Message: "Nama tidak terdeteksi"})
		}
	case "visa":
		if extracted.NoVisa == "" {
			errors = append(errors, ValidationError{Field: "no_visa", Message: "Nomor visa tidak terdeteksi"})
		}
	}

	return errors
}

func normalizeToSiskopatuh(data any, docType string) any {
	extracted, ok := data.(ExtractedFields)
	if !ok {
		return data
	}

	normalized := map[string]any{}

	// Nama and Nama Paspor must both carry the same name, but the OCR sometimes
	// fills only one of them (e.g. a passport's name read into nama_paspor only).
	// Resolve one name (prefer Nama, fall back to NamaPaspor) and mirror it into
	// both keys so neither column comes out blank.
	if name := extracted.Nama; name != "" || extracted.NamaPaspor != "" {
		if name == "" {
			name = extracted.NamaPaspor
		}
		normalized["nama"] = name
		normalized["nama_paspor"] = name
	}
	// Identity number = NIK for a KTP, but the PASSPORT NUMBER for a passport
	// (a passport has no NIK). Jenis Identitas uses the Siskopatuh dropdown
	// values NIK / PASPOR (template Sheet2 col E), not "KTP"/"Paspor".
	if extracted.NIK != "" {
		normalized["no_identitas"] = extracted.NIK
		normalized["jenis_identitas"] = "NIK"
	} else if extracted.NoPaspor != "" {
		normalized["no_identitas"] = extracted.NoPaspor
		normalized["jenis_identitas"] = "PASPOR"
	}
	if extracted.NoPaspor != "" {
		normalized["no_paspor"] = extracted.NoPaspor
	}
	if extracted.TempatLahir != "" {
		normalized["tempat_lahir"] = extracted.TempatLahir
	}
	if extracted.TanggalLahir != "" {
		normalized["tanggal_lahir"] = extracted.TanggalLahir
	}
	if extracted.JenisKelamin != "" {
		normalized["gender"] = extracted.JenisKelamin
	}
	if extracted.Alamat != "" {
		normalized["alamat"] = extracted.Alamat
	}
	canonProv := mapProvinsi(extracted.Provinsi)
	if canonProv != "" {
		normalized["provinsi"] = canonProv
	}
	if extracted.Kabupaten != "" {
		normalized["kabupaten"] = mapKabupaten(canonProv, extracted.Kabupaten)
	}
	if extracted.Kecamatan != "" {
		normalized["kecamatan"] = extracted.Kecamatan
	}
	if extracted.Kelurahan != "" {
		normalized["kelurahan"] = extracted.Kelurahan
	}
	if extracted.NoTelepon != "" {
		normalized["no_telepon"] = extracted.NoTelepon
	}
	if extracted.NoHP != "" {
		normalized["no_hp"] = extracted.NoHP
	}
	if kw := mapKewarganegaraan(extracted.Kewarganegaraan); kw != "" {
		normalized["kewarganegaraan"] = kw
	}
	if extracted.GolonganDarah != "" {
		normalized["golongan_darah"] = extracted.GolonganDarah
	}
	if pd := mapPendidikan(extracted.Pendidikan); pd != "" {
		normalized["pendidikan"] = pd
	}
	if pk := mapPekerjaan(extracted.Pekerjaan); pk != "" {
		normalized["pekerjaan"] = pk
	}
	if extracted.StatusPerkawinan != "" {
		normalized["status_pernikahan"] = mapStatusNikah(extracted.StatusPerkawinan)
	}
	if extracted.Agama != "" {
		normalized["agama"] = extracted.Agama
	}

	if extracted.TanggalPaspor != "" {
		normalized["tanggal_paspor"] = extracted.TanggalPaspor
	}
	if extracted.TanggalExpired != "" {
		normalized["tanggal_expired_paspor"] = extracted.TanggalExpired
	}
	if extracted.KotaPaspor != "" {
		normalized["kota_paspor"] = extracted.KotaPaspor
	}

	if pv := mapProviderVisa(extracted.ProviderVisa); pv != "" {
		normalized["provider_visa"] = pv
	}
	if extracted.NoVisa != "" {
		normalized["no_visa"] = extracted.NoVisa
	}
	if extracted.TanggalVisa != "" {
		normalized["tanggal_visa"] = extracted.TanggalVisa
	}
	if extracted.TanggalVisaAkhir != "" {
		normalized["tanggal_visa_akhir"] = extracted.TanggalVisaAkhir
	}

	normalized["source_doc_type"] = docType
	normalized["siskopatuh_version"] = "2.0"

	return normalized
}
