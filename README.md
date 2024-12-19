# **API Documentation: Media Service**

## **Endpoint**

### **POST** `/api/v1/auth/media/upload`

## **Headers**

| Header          | Value               | Description                                       |
| --------------- | ------------------- | ------------------------------------------------- |
| `Authorization` | Bearer `<token>`    | Token autentikasi pengguna (wajib).               |
| `Content-Type`  | multipart/form-data | Header untuk mendukung pengunggahan file (wajib). |

---

## **Request Body**

Body harus dikirimkan dalam format `multipart/form-data`.

| Parameter | Type     | Description                                   |
| --------- | -------- | --------------------------------------------- |
| `file`    | `file`   | File gambar yang akan diunggah (wajib).       |
| `folder`  | `string` | Nama folder tujuan di penyimpanan (opsional). |

---

### **Example Request**

#### **Curl Command**

```bash
curl --location --request POST 'https://yourdomain.com/api/v1/auth/media/upload' \
--header 'Authorization: Bearer your_access_token' \
--form 'file=@"/path/to/your/image.jpg"' \
--form 'folder="profile_pictures"'
```

---

## **Responses**

### **Success Response**

**HTTP Status Code:** `200 OK`

```json
{
  "status": "success",
  "message": "File uploaded successfully",
  "data": {
    "url": "https://yourstorage.com/path/to/image.jpg",
    "public_id": "folder_name/image_id",
    "size": "1024KB",
    "type": "image/jpeg"
  }
}
```

| Field       | Type     | Description                               |
| ----------- | -------- | ----------------------------------------- |
| `url`       | `string` | URL untuk mengakses gambar yang diunggah. |
| `public_id` | `string` | ID publik file di penyimpanan.            |
| `size`      | `string` | Ukuran file gambar.                       |
| `type`      | `string` | Tipe MIME dari file yang diunggah.        |

---

### **Error Responses**

#### **400 Bad Request**

**Reason:** File tidak ditemukan atau format file tidak didukung.

```json
{
  "status": "fail",
  "message": "File is required"
}
```

#### **401 Unauthorized**

**Reason:** Token autentikasi tidak valid atau tidak ada.

```json
{
  "status": "fail",
  "message": "Unauthorized. Please provide a valid token."
}
```

#### **500 Internal Server Error**

**Reason:** Terjadi kesalahan di server.

```json
{
  "status": "error",
  "message": "An error occurred while processing the file."
}
```

---

## **Validation Rules**

- **File wajib**: Parameter `file` harus diberikan.
- **Format didukung**: File harus berupa salah satu tipe berikut:
  - `image/jpeg`
  - `image/png`
- **Ukuran maksimum**: File tidak boleh melebihi 5MB (atau sesuai batas server).

---

## **Notes**

Pastikan token autentikasi (`Bearer Token`) valid sebelum mengakses endpoint ini. Folder default untuk penyimpanan adalah `uploads` jika parameter `folder` tidak diberikan.

---
