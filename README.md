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

#### **401 Unauthorized**

**Reason:** Token autentikasi tidak valid atau tidak ada.

```json
{
  "status": "fail",
  "message": "Unauthorized. Please provide a valid token."
}
```

## **Notes**

Pastikan token autentikasi (`Bearer Token`) valid sebelum mengakses endpoint ini.

---
