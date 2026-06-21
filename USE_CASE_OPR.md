# Use Case: SMS — Safety Management System

**Studi kasus:** Kabel listrik terkelupas di area gudang penyimpanan, PT Yusen Logistics Interlink Indonesia (CL Group)

Dokumen ini menelusuri satu laporan risiko keselamatan dari awal (ditemukan oleh karyawan) sampai tuntas (ditutup oleh Head of Department), agar siapa pun bisa memahami cara kerja aplikasi dari sudut pandang pengguna sehari-hari.

---

## Pemain dalam kasus ini

| Peran | Nama (contoh) | Tanggung jawab di alur ini |
|---|---|---|
| Pelapor | Deni (Staf Gudang) | Menemukan & melaporkan risiko |
| PIC (Person in Charge) | Ratna (Staff Safety, Warehouse Dept) | Mengklasifikasi laporan, menginvestigasi penyebab, mengusulkan solusi |
| HOD (Head of Department) | Pak Reza (Warehouse Manager) | Mereview dan menyetujui atau menolak hasil investigasi |

---

## Tahap 0 — Konteks

Jam 14.30, Deni sedang memindahkan palet di **Gudang B, BTLC Site**. Di sudut dekat rak penyimpanan barang, dia melihat kabel listrik yang terkelupas di dekat lantai — kabel itu tampaknya terjepit rak besi pada saat rak dipindahkan beberapa waktu lalu, dan isolasinya terkelupas sepanjang kira-kira 15 cm. Area di sekitar kabel itu sering dilalui staf dan sesekali basah karena dekat dengan titik kebocoran atap yang sudah lama dilaporkan ke bagian gedung.

Deni menyadari ini berpotensi menyebabkan sengatan listrik atau kebakaran. Di dinding dekat rak tersebut ada **stiker QR code SMS** yang terpasang di setiap area kerja.

---

## Tahap 1 — Deklarasi (oleh Deni)

**Siapa:** Deni — tidak perlu login, siapa pun bisa melapor
**Cara akses:** Scan QR code di area gudang menggunakan HP

### Langkah yang dilakukan Deni:

1. Deni scan QR code di dekat rak penyimpanan Gudang B
2. Form laporan langsung terbuka di HP — satu halaman, bisa diisi sambil berdiri di lokasi
3. Deni mengisi:
   - **Tingkat Risiko:** Deni pilih **High** — karena membaca deskripsi di bawah setiap pilihan, "kabel terbuka di area yang bisa terkena air" masuk kategori "dapat mengancam keselamatan dalam waktu dekat"
   - **Deskripsi:** *"Kabel listrik terkelupas di sudut Rak B-12, dekat lantai. Isolasi rusak sekitar 15 cm, kemungkinan terjepit rak. Area sekitar kadang basah karena atap bocor. Berbahaya jika disentuh atau terkena air."*
   - **Lokasi:** Sudah terisi otomatis dari QR code → "Gudang B · BTLC"
   - **Departemen:** Sudah terisi otomatis → "Warehouse (WHD)"
   - **Nama Pelapor:** Deni mengisi namanya agar mudah dihubungi jika ada pertanyaan (bisa dikosongkan jika ingin anonim)
   - **Foto:** Deni ambil foto kabel yang rusak dari jarak aman — foto otomatis mendapat cap waktu dan nama pelapor
4. Deni tekan **Kirim Laporan**

### Yang dilihat Deni setelah kirim:

```
✓ Laporan Terkirim
Kode Laporan: R000094
Simpan kode ini untuk referensi
```

Deni catat kode tersebut, lalu segera memberitahu rekan-rekannya untuk tidak mendekati area Rak B-12 sambil menunggu penanganan.

---

## Tahap 2 — Klasifikasi & Konfirmasi (oleh Ratna, PIC)

**Siapa:** Ratna — login ke aplikasi SMS dengan akun kantornya
**SLA:** Karena tingkat risiko High, Ratna punya waktu **4 jam kerja** untuk mengkonfirmasi klasifikasi

### Yang dilakukan Ratna:

1. Ratna login ke SMS dan langsung melihat notifikasi laporan baru di **Dashboard** dengan badge "Declaration"
2. Ratna buka laporan **R000094**, membaca deskripsi dan melihat foto dari Deni
3. Ratna mengisi form Klasifikasi:
   - **Tingkat Risiko:** Ratna konfirmasi tetap **High** — sesuai penilaian Deni
   - **Prioritas:** Ratna set **High** — situasi harus ditangani segera
   - **Catatan:** *"Dikonfirmasi — kabel terbuka terlihat jelas di foto. Mengingat atap di area tersebut diketahui bocor, ini risiko kesetrum dan kebakaran yang tidak bisa ditunda. Area sementara dikunci akses sambil menunggu investigasi."*
4. Ratna tekan **Confirm Classification**

### Yang dilihat Ratna:

Alur laporan di atas halaman (stepper) bergerak maju — tahap "Classification" sudah selesai, sekarang masuk ke tahap "Investigation". Ratna juga langsung turun ke lapangan untuk memulai pengecekan.

---

## Tahap 3 — Investigasi & Solusi (oleh Ratna, PIC)

**Siapa:** Ratna
**SLA:** 8 jam kerja dari waktu deklarasi (High priority)

### Yang dilakukan Ratna:

Ratna berkoordinasi dengan tim teknik (Engineering & Maintenance) dan tim pengelola gedung (GA/Facility).

1. Ratna kembali ke halaman laporan R000094 dan mengisi form Investigasi:
   - **Akar Masalah:**
     *"Kabel instalasi listrik bagian rak B-12 tertindih sudut besi rak pada saat proses relayout gudang 3 minggu lalu. Tidak ada pengecekan kondisi instalasi listrik setelah proses pemindahan rak selesai. Kondisi diperparah oleh kebocoran atap di titik yang berdekatan — air menggenang sesekali di lantai dekat kabel tersebut."*
   - **Langkah Penanganan:**
     *"1) Area Rak B-12 langsung dipasang garis pembatas dan rambu 'BAHAYA — DILARANG MASUK' per hari ini. 2) Tim Engineering sudah dipanggil dan dijadwalkan perbaikan kabel hari ini pukul 16.00. 3) Kebocoran atap di titik yang sama dilaporkan ke tim GA untuk ditangani bersamaan. 4) Diusulkan agar setiap proses pemindahan rak atau relayout gudang ke depannya wajib disertai pengecekan kondisi instalasi listrik oleh tim Engineering sebelum area dibuka kembali."*
2. Ratna upload foto tambahan — foto rambu pembatas yang sudah dipasang, dan foto kondisi kabel setelah tim Engineering selesai memperbaiki
3. Ratna tekan **Submit Investigation**

### Yang dilihat Ratna:

Tahap "Investigation" sudah selesai di stepper. Laporan kini menunggu review dari Pak Reza sebagai HOD.

---

## Tahap 4 — Persetujuan (oleh Pak Reza, HOD)

**Siapa:** Pak Reza — login ke SMS dengan akun HOD
**SLA:** 10 jam kerja dari waktu deklarasi (High priority)

### Skenario A — Pak Reza menyetujui (jalur normal)

1. Pak Reza membuka laporan R000094, menelusuri seluruh riwayat: laporan awal Deni, klasifikasi Ratna, hasil investigasi, hingga foto-foto sebelum dan sesudah perbaikan
2. Pak Reza puas dengan penanganan yang dilakukan. Dia menambahkan komentar:
   *"Penanganan cepat dan tepat. Saya setuju dengan usulan SOP baru untuk pengecekan instalasi listrik setiap kali ada relayout gudang. Tolong koordinasikan penyusunan SOP-nya dengan tim HSE bulan ini."*
3. Pak Reza klik **Approve**

**Hasilnya:** Laporan R000094 resmi berstatus **Closed**. Seluruh riwayat tersimpan dan bisa dibuka kapan pun.

---

### Skenario B — Pak Reza meminta revisi

Misalkan Pak Reza merasa ada yang kurang — perbaikan kabel memang sudah selesai, tapi dia ingin ada bukti bahwa kebocoran atap juga sudah ditangani sebelum laporan ditutup.

1. Pak Reza mengisi komentar (**wajib diisi jika menolak**):
   *"Perbaikan kabel sudah bagus, tapi kebocoran atap di area yang sama belum ada bukti penanganannya. Mohon dilampirkan konfirmasi dari tim GA bahwa kebocoran sudah diperbaiki atau minimal sudah ada jadwal perbaikan resmi. Baru bisa saya tutup laporannya."*
2. Pak Reza klik **Reject**

**Hasilnya:** Laporan kembali ke tahap investigasi. Ratna mendapat notifikasi dan diminta melengkapi data, lalu submit ulang. Setelah Ratna melampirkan konfirmasi dari tim GA, Pak Reza kembali mereview dan kali ini menyetujui.

---

## Tahap 5 — Selesai (Closed)

Setelah Pak Reza menyetujui, laporan R000094 berstatus **Closed**.

Semua orang yang membuka laporan ini ke depannya akan melihat:
- Seluruh kronologi dari awal hingga akhir
- Foto kondisi sebelum dan sesudah perbaikan
- Catatan dari setiap pihak yang terlibat
- Laporan ini juga masuk ke rekap statistik bulanan perusahaan

---

## Yang terjadi jika tenggat waktu terlewat

Andaikan Ratna terlambat menyelesaikan investigasi melewati batas SLA 8 jam:

- Sistem secara otomatis menandai laporan R000094 sebagai **Overdue**
- Badge merah **"Terlambat"** muncul di daftar laporan — terlihat oleh semua pihak yang berwenang
- Email peringatan otomatis dikirim

---

## Ringkasan alur

```
Deni (staf lapangan)      Ratna (PIC)                    Pak Reza (HOD)
        │                      │                                │
        ├─ Temukan kabel rusak  │                                │
        ├─ Scan QR + isi form   │                                │
        ├─ Kirim laporan        │                                │
        │   → Kode: R000094     │                                │
        │                       ├─ Baca laporan & foto           │
        │                       ├─ Konfirmasi tingkat risiko     │
        │                       ├─ Pasang pembatas area          │
        │                       ├─ Investigasi + koordinasi tim  │
        │                       ├─ Upload bukti foto perbaikan   │
        │                       ├─ Submit investigasi            │
        │                       │                                │
        │                       │                  ├─ Review semua riwayat
        │                       │                  ├─ Approve → CLOSED ✓
        │                       │◄── Reject (jika ada kekurangan)
        │                       ├─ Lengkapi data & submit ulang  │
        │                       │                  └─ Approve → CLOSED ✓
```

---

## Hal penting yang perlu diingat

- **Siapa pun bisa melapor** — tidak perlu akun, cukup scan QR code di area kerja. Nama pelapor pun boleh dikosongkan jika ingin anonim
- **Kode laporan** (contoh: R000094) adalah cara tercepat untuk melacak status laporan — simpan kode ini setiap kali selesai melapor
- **Setiap tahap ada tenggat waktu otomatis** — semakin tinggi tingkat risiko, semakin singkat waktu yang diberikan untuk menanganinya
- **Semua riwayat tersimpan** — foto, catatan, dan siapa yang melakukan apa, semuanya bisa dilihat kembali kapan pun
- **Jika HOD menolak**, bukan berarti laporan gagal — itu artinya perlu dilengkapi dulu sebelum bisa ditutup secara resmi