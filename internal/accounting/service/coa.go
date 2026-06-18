package service

import "github.com/jamaah-in/v2/internal/accounting/model"

// Standard account codes (Bagan Akun standar travel umroh/haji). Kept as
// constants so the posting engine and the seed never drift.
const (
	AccKas               = "1101" // Kas
	AccBank              = "1102" // Bank
	AccPiutangJemaah     = "1201" // Piutang Jemaah (AR)
	AccPersediaan        = "1301" // Persediaan Perlengkapan
	AccHutangVendor      = "2101" // Hutang Vendor (AP)
	AccHutangKomisi      = "2102" // Hutang Komisi Agen
	AccHutangGaji        = "2103" // Hutang Gaji
	AccHutangPajak       = "2104" // Hutang Pajak (PPh/BPJS)
	AccHutangTabungan    = "2201" // Hutang Tabungan Jemaah
	AccTitipanJemaah     = "2202" // Hutang Titipan/Uang Muka Jemaah (kelebihan bayar)
	AccModal             = "3101" // Modal Disetor
	AccSaldoAwal         = "3901" // Saldo Awal (opening balance equity)
	AccPendapatanPaket   = "4101" // Pendapatan Paket
	AccPendapatanLain    = "4901" // Pendapatan Lain-lain (mis. kas lebih)
	AccBebanKomisi       = "5101" // Beban Komisi Agen
	AccBebanGaji         = "5102" // Beban Gaji
	AccBebanPerlengkapan = "5103" // Beban Perlengkapan
	AccBebanOperasional  = "5901" // Beban Operasional Lain
	AccBebanSelisihKas   = "5902" // Beban Selisih Kas (kas kurang)
)

// StandardCOA is the default chart of accounts seeded per org on first use.
func StandardCOA() []model.Account {
	a := func(code, name, typ string) model.Account {
		nb := model.BalanceCredit
		if typ == model.TypeAsset || typ == model.TypeExpense {
			nb = model.BalanceDebit
		}
		return model.Account{Code: code, Name: name, Type: typ, NormalBalance: nb, IsActive: true}
	}
	return []model.Account{
		a(AccKas, "Kas", model.TypeAsset),
		a(AccBank, "Bank", model.TypeAsset),
		a(AccPiutangJemaah, "Piutang Jemaah", model.TypeAsset),
		a(AccPersediaan, "Persediaan Perlengkapan", model.TypeAsset),
		a(AccHutangVendor, "Hutang Vendor", model.TypeLiability),
		a(AccHutangKomisi, "Hutang Komisi Agen", model.TypeLiability),
		a(AccHutangGaji, "Hutang Gaji", model.TypeLiability),
		a(AccHutangPajak, "Hutang Pajak", model.TypeLiability),
		a(AccHutangTabungan, "Hutang Tabungan Jemaah", model.TypeLiability),
		a(AccTitipanJemaah, "Hutang Titipan Jemaah", model.TypeLiability),
		a(AccModal, "Modal Disetor", model.TypeEquity),
		a(AccSaldoAwal, "Saldo Awal", model.TypeEquity),
		a(AccPendapatanPaket, "Pendapatan Paket", model.TypeRevenue),
		a(AccPendapatanLain, "Pendapatan Lain-lain", model.TypeRevenue),
		a(AccBebanKomisi, "Beban Komisi Agen", model.TypeExpense),
		a(AccBebanGaji, "Beban Gaji", model.TypeExpense),
		a(AccBebanPerlengkapan, "Beban Perlengkapan", model.TypeExpense),
		a(AccBebanOperasional, "Beban Operasional Lain", model.TypeExpense),
		a(AccBebanSelisihKas, "Beban Selisih Kas", model.TypeExpense),
	}
}
