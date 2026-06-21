# Use Case: SMS — Safety Management System

**Studi kasus:** Tumpahan oli di area loading dock, PT Yusen Logistics Interlink Indonesia (CL Group)

Dokumen ini menelusuri satu laporan risiko keselamatan dari awal (deklarasi oleh karyawan lapangan) sampai akhir (ditutup oleh Head of Department), supaya kamu paham alur penggunaan aplikasi secara end-to-end.

---

## Pemain dalam kasus ini

| Peran | Nama (contoh) | Tanggung jawab di alur ini |
|---|---|---|
| Pelapor | Budi (Forklift Operator) | Menemukan & melaporkan risiko |
| PIC (Person in Charge) | Siti (Staff Safety, Warehouse Dept) | Klasifikasi laporan, investigasi penyebab, usulkan solusi |
| HOD (Head of Department) | Pak Andi (Warehouse Manager) | Review & approve/reject hasil investigasi |

---

## Tahap 0 — Konteks

Jam 09:15, Budi sedang mengoperasikan forklift di **Loading Dock B, BTLC Site**. Dia melihat ada tumpahan oli hidrolik sekitar 1 meter persegi di jalur lalu lalang forklift, dekat pintu masuk dock. Lantai licin, belum ada rambu peringatan, dan jalur itu cukup ramai dilalui staff lain.

Budi tahu ini risiko kecelakaan — bisa bikin orang lain terpeleset atau forklift selip. Di dinding dekat dock itu ada **stiker QR code SMS** yang memang dipasang di tiap area kerja sesuai SOP.

---

## Tahap 1 — Declaration (Deklarasi oleh Budi)

**Siapa:** Budi (tidak perlu akun/login — form deklarasi bersifat publik)
**Halaman:** Form deklarasi mobile (hasil scan QR)
**Endpoint backend:** `POST /api/reports`

### Langkah yang dilakukan Budi:

1. Budi scan QR code di dekat Loading Dock B pakai HP-nya
2. Form terbuka, otomatis ke halaman **Buat Laporan Risiko** (mobile-first, satu scroll)
3. Budi isi:
   - **Tingkat Risiko:** Pilih **Medium** — karena di bawah deskripsi tiap pilihan, dia baca "Ada risiko terjadinya kecelakaan", itu paling sesuai (bukan "mengancam saat ini juga" karena belum ada yang terpeleset, tapi juga bukan cuma "jika tidak segera diperbaiki")
   - **Deskripsi:** "Ada tumpahan oli hidrolik di jalur forklift dekat pintu masuk Loading Dock B, sekitar 1x1 meter. Lantai licin, belum ada rambu peringatan. Jalur ini sering dilalui staff dan forklift."
   - **Lokasi:** Sudah ter-prefill otomatis dari QR code → "Loading Dock B · BTLC"
   - **Departemen:** Sudah ter-prefill otomatis → "Warehouse (WHD)"
   - **Nama Pelapor:** Budi isi namanya (opsional — dia bisa juga kosongkan untuk anonim, tapi dia pilih cantumkan nama supaya gampang dihubungi kalau ada pertanyaan)
   - **Foto:** Budi ambil foto langsung pakai kamera HP — otomatis kena watermark timestamp + nama, supaya jelas kapan dan oleh siapa foto itu diambil
4. Budi tekan **Kirim Laporan**

### Yang terjadi di balik layar:

- Backend (`ReportService.Declare`) generate kode laporan otomatis: **R000087**
- Status laporan: `declaration`
- File foto Budi tersimpan di `report_files`, terhubung ke laporan ini
- Sebuah baris dicatat di `report_activities`: "declared" oleh "Budi"
- Email notifikasi coba dikirim — tapi karena **PIC belum di-assign** di tahap ini, sesuai SOP, notifikasi awal ini sebenarnya baru efektif setelah ada PIC (ini bagian yang masih perlu disempurnakan di backend, sudah kita bahas sebelumnya)

### Yang dilihat Budi:

Layar sukses muncul:
```
✓ Laporan Terkirim
Kode Laporan: R000087
Simpan kode ini untuk referensi
```

Budi foto layar ini sebagai bukti dia sudah lapor, lalu lanjut kerja.

---

## Tahap 2 — Classification & Confirmation (oleh Siti, PIC)

**Siapa:** Siti (role: `pic`)
**Halaman:** `/reports/87` (halaman detail, kolom tengah — form "Classify & Confirm")
**Endpoint backend:** `PATCH /api/reports/87/classify`
**SLA:** Karena risk level akhirnya ditetapkan Medium → **8 jam kerja** dari waktu deklarasi untuk konfirmasi klasifikasi

### Yang terjadi:

1. Siti login ke SMS pakai username/password kantornya
2. Di **Dashboard**, dia lihat ada laporan baru masuk dengan status "Declaration"
3. Siti klik tab **Reports**, cari kode **R000087** di tabel, klik row-nya
4. Halaman detail terbuka — Siti lihat **stepper di atas**: titik pertama ("Declaration") sudah hijau/centang, titik kedua ("Classification") sedang aktif
5. Di kolom tengah ada form **Classify & Confirm** (karena Siti adalah PIC dan status masih `declaration`, sistem otomatis menampilkan form ini ke dia)
6. Siti baca deskripsi & foto Budi di kolom kanan (section Files), lalu isi:
   - **Risk Level:** Konfirmasi tetap **Medium**
   - **Priority:** Siti set **Medium** juga (ini akan menentukan SLA tahap berikutnya: 16 jam untuk investigasi, 18 jam untuk approval)
   - **Notes:** "Dikonfirmasi sesuai foto — tumpahan oli area dock B, perlu segera dibersihkan dan dipasang rambu sementara sambil diinvestigasi penyebabnya."
7. Siti tekan **Confirm Classification**

### Yang terjadi di balik layar:

- Backend (`ReportService.Classify`) validasi: status harus `declaration` — sesuai, lanjut
- Status berubah: `declaration` → `classification`
- `PICID` di-set ke ID Siti (di real implementation, field PIC assignment perlu ditambahkan ke form — saat ini classify endpoint menerima `pic_id` opsional)
- SLA deadline baru dihitung otomatis: `SLADeadlineInvestigation` dan `SLADeadlineApproval` di-set berdasarkan tabel SLA Medium (16 jam & 18 jam dari sekarang)
- Activity log baru: "classified" oleh Siti, dari status `declaration` ke `classification`
- Email notifikasi terkirim ke Siti sendiri (sebagai PIC) — "Update on Risk Report R000087 - Classification confirmed"

### Yang dilihat Siti:

Stepper berubah — titik kedua ("Classification") sekarang juga hijau, titik ketiga ("Investigation") jadi aktif.

---

## Tahap 3 — Causes & Countermeasures (Investigasi oleh Siti)

**Siapa:** Siti (role: `pic`)
**Halaman:** `/reports/87` (kolom tengah — form "Causes & Countermeasures")
**Endpoint backend:** `PATCH /api/reports/87/investigate`
**SLA:** 16 jam kerja dari waktu deklarasi (Medium priority)

### Yang terjadi:

Siti turun ke lapangan, cek selang hidrolik forklift area dock B, dan ngobrol dengan tim maintenance.

1. Siti balik ke halaman detail laporan R000087
2. Karena status sekarang `classification`, form **Causes & Countermeasures** otomatis muncul di kolom tengah
3. Siti isi:
   - **Root Cause:** "Selang hidrolik forklift unit FL-12 mengalami keausan di sambungan, menyebabkan rembesan oli saat unit diparkir. Belum ada jadwal preventive maintenance untuk pengecekan selang secara berkala."
   - **Countermeasures:** "1) Selang hidrolik FL-12 sudah diganti hari ini. 2) Area tumpahan sudah dibersihkan dan diberi rambu. 3) Diusulkan menambahkan item pengecekan kondisi selang hidrolik ke checklist preventive maintenance mingguan untuk semua unit forklift."
4. Siti juga upload foto bukti — selang baru yang sudah dipasang, dan area dock yang sudah bersih (lewat section **Files** di kolom kanan, upload tambahan)
5. Siti tekan **Submit Investigation**

### Yang terjadi di balik layar:

- Backend validasi: status harus `classification` atau `investigation` — sesuai
- Status berubah: `classification` → `investigation`
- `RootCause` dan `Countermeasures` tersimpan
- Activity log: "investigated" oleh Siti
- Email terkirim — "Review Required: Risk Report R000087" (harusnya ke Pak Andi sebagai HOD, sesuai catatan kita sebelumnya soal `getRecipients()` yang masih perlu diperbaiki supaya kirim ke HOD juga, bukan cuma PIC)

### Yang dilihat Siti:

Stepper: titik ketiga ("Investigation") jadi hijau, titik keempat ("Approval") jadi aktif. Tugas Siti di laporan ini selesai — sekarang menunggu Pak Andi.

---

## Tahap 4 — Verification and Approval (oleh Pak Andi, HOD)

**Siapa:** Pak Andi (role: `hod`)
**Halaman:** `/reports/87` (kolom tengah — section "HOD Review")
**Endpoint backend:** `POST /api/reports/87/approve` atau `POST /api/reports/87/reject`
**SLA:** 18 jam kerja dari waktu deklarasi (Medium priority)

### Skenario A — Pak Andi setuju (jalur normal)

1. Pak Andi login, buka laporan R000087 dari tab Reports
2. Dia baca seluruh riwayat: deskripsi awal Budi, klasifikasi Siti, root cause, countermeasures, dan foto-foto bukti (sebelum/sesudah) di kolom kanan
3. Karena status sekarang `investigation` dan dia adalah HOD, section **HOD Review** muncul dengan tombol **Approve** dan **Reject**
4. Pak Andi puas dengan penjelasan dan bukti foto, isi comment (opsional): "Penanganan cepat dan tepat. Setuju usulan tambahan checklist maintenance mingguan — tolong koordinasikan dengan tim Asset Management."
5. Pak Andi klik **Approve**

**Yang terjadi di balik layar:**
- Backend (`ReportService.Approve`) validasi status `investigation` — sesuai
- Status berubah: `investigation` → `closed` langsung (tidak ada status `approval` terpisah di backend — itu cuma representasi visual di stepper, di balik layar approve langsung menutup laporan)
- Activity log: "approved" oleh Pak Andi
- Email "Closed: Risk Report R000087" terkirim

**Yang dilihat tim:** Stepper penuh — semua 5 titik hijau, termasuk "Closed". Laporan selesai.

---

### Skenario B — Pak Andi tidak puas (jalur revisi)

Misalkan dalam kasus ini Pak Andi merasa countermeasures Siti belum cukup — dia mau ada bukti training ulang operator forklift, bukan cuma ganti selang.

1. Pak Andi isi comment (**wajib** untuk reject): "Bukti penggantian selang sudah cukup, tapi mohon dilengkapi juga rencana refresher training untuk operator forklift terkait deteksi dini kebocoran hidrolik. Mohon update dan submit ulang."
2. Pak Andi klik **Reject**

**Yang terjadi di balik layar:**
- Status berubah: `investigation` → `classification` (balik mundur satu langkah, sesuai SOP "sistem otomatis kembali ke step Cause & Countermeasures")
- Email "Rejected: Risk Report R000087 - Revision Required" terkirim ke Siti
- Stepper otomatis menunjukkan titik "Investigation" kembali jadi belum-selesai

**Yang dilakukan Siti:** Dia buka lagi laporan R000087, lihat comment Pak Andi di kolom tengah (section "HOD Feedback"), lengkapi root cause/countermeasures dengan rencana training, lalu **Submit Investigation** ulang — alur kembali ke Tahap 4.

---

## Tahap 5 — Closed

Setelah Pak Andi approve (baik di percobaan pertama atau setelah revisi), laporan R000087 berstatus **Closed**.

Yang bisa dilihat siapa saja yang membuka laporan ini setelahnya:
- Stepper lengkap 5 centang hijau
- Seluruh riwayat di **Activity Log** (kolom kanan): declared → classified → investigated → (mungkin rejected →) investigated lagi → approved
- Semua foto: foto awal Budi, foto bukti perbaikan Siti
- Dashboard perusahaan (`/dashboard`) ikut ter-update: laporan ini masuk hitungan "Closed" di pie chart Processing Status, dan masuk grafik Monthly Statistics bulan berjalan dengan kategori risk Medium

---

## Yang terjadi kalau SLA terlewat (skenario tambahan)

Misalkan Siti lupa submit investigasi dan 16 jam kerja terlewat tanpa update:

- **Scheduler** (`SLAChecker`, jalan tiap interval sesuai `.env`) otomatis mendeteksi laporan R000087 melewati `SLADeadlineInvestigation`
- Field `IsOverdue` di-set `true`
- Email **"OVERDUE ALERT: Risk Report R000087"** otomatis terkirim
- Di tabel **Reports** dan halaman detail, badge merah **"Overdue"** muncul di sebelah kode laporan — siapapun yang lihat tabel langsung tahu ada laporan yang telat

---

## Ringkasan alur (versi singkat)

```
Budi (lapangan)          Siti (PIC)                    Pak Andi (HOD)
     │                        │                              │
     ├─ Scan QR               │                              │
     ├─ Isi form + foto       │                              │
     ├─ Submit ──► R000087    │                              │
     │   status: declaration  │                              │
     │                        ├─ Buka laporan                │
     │                        ├─ Konfirmasi risk + priority   │
     │                        ├─ Confirm ──► status: classification
     │                        │                              │
     │                        ├─ Investigasi lapangan         │
     │                        ├─ Isi root cause + solusi      │
     │                        ├─ Upload bukti foto             │
     │                        ├─ Submit ──► status: investigation
     │                        │                              │
     │                        │                              ├─ Review semua data
     │                        │                              ├─ Approve atau Reject
     │                        │                              │
     │                        │            ┌─ Approve ──► status: closed ✓
     │                        │◄─ Reject ───┤
     │                        │             (balik ke classification, ulangi investigasi)
```

---

## Catatan untuk yang baru pertama kali pakai

- **Tidak perlu akun** untuk melaporkan risiko — siapa saja bisa scan QR dan lapor, bahkan anonim
- **Login dibutuhkan** mulai dari tahap Classification ke atas (PIC, HOD, Admin)
- **Role menentukan apa yang bisa dilakukan**: PIC hanya bisa classify & investigate, HOD hanya bisa approve/reject, Admin bisa semua + kelola Master Data (Users, Departments, Locations)
- **Kode laporan** (format `R000XXX`) adalah cara tercepat untuk mencari laporan tertentu — simpan kode ini setiap kali lapor
- **SLA berjalan otomatis** berdasarkan priority yang ditentukan PIC saat classify — semakin tinggi priority, semakin ketat tenggat waktunya