# 🔐 Secure Authentication Service (Go + Redis + PostgreSQL)

A production-ready authentication backend built with **Golang**, designed to handle **secure login**, **brute-force protection**, and **timing attack mitigation** using modern backend practices.

---

## 🚀 Features

- 🔑 Secure user registration with **bcrypt hashing**
- 🔐 Login system with **timing attack protection**
- 🚫 Account lockout after multiple failed attempts
- 🐢 Progressive delay to slow brute-force attacks
- 🌍 IP-based rate limiting
- ⚡ Redis-powered high-performance tracking
- 🧠 Protection against user enumeration attacks
- 🧱 Clean and scalable architecture using Chi router

---

## 🏗️ Tech Stack

- **Go (Golang)** – Backend server
- **PostgreSQL** – Database
- **Redis** – Rate limiting & lock management
- **Chi Router** – HTTP routing
- **bcrypt** – Password hashing

---

## 📂 Project Structure


.
├── main.go
├── db/
│ ├── queries.sql
│ └── generated files (sqlc)
├── .env
└── README.md


---

## ⚙️ Environment Variables

Create a `.env` file in the root directory:


POSTGRES_CONN=postgres://user:password@localhost:5432/dbname
REDIS_CONN=redis://localhost:6379


---

## 🛠️ Installation & Setup

### 1. Clone the repository

git clone <your-repo-url>
cd project


### 2. Install dependencies

go mod tidy


### 3. Start required services
Ensure the following are running:
- PostgreSQL
- Redis

### 4. Run the server

go run main.go


Server will start at:

http://localhost:8080


---

## 📡 API Endpoints

### 🔹 Register

**POST** `/register`

#### Request Body:

{
"username": "testuser",
"password": "securepassword"
}


#### Response:

201 Created


---

### 🔹 Login

**POST** `/login`

#### Request Body:

{
"username": "testuser",
"password": "securepassword"
}


#### Responses:

200 OK → Login successful
401 Unauthorized → Invalid username or password
403 Forbidden → Account locked
429 Too Many Requests → Rate limit exceeded


---

## 🔐 Security Mechanisms

### 1. Password Hashing
- Uses `bcrypt`
- Passwords are never stored in plain text

---

### 2. Timing Attack Protection
- Uses `bcrypt.CompareHashAndPassword`
- Prevents attackers from guessing passwords via response timing

---

### 3. Failed Attempt Tracking (Redis)

#### Redis Keys:

login:attempts:user:<userID>
login:attempts:ip:<ip>
login:lock:user:<userID>


---

### 4. Account Locking

- Locks account after **5 failed attempts**
- Lock duration: **15 minutes**

---

### 5. Progressive Delay

Each failed login increases response delay:

| Attempts | Delay |
|--------|------|
| 1 | 500ms |
| 2 | 1s |
| 3 | 1.5s |
| 4+ | max 5s |

---

### 6. IP Rate Limiting

- Tracks requests per IP
- Blocks after **20 attempts within 15 minutes**

---

### 7. User Enumeration Protection

All authentication errors return:

Invalid username or password


This prevents attackers from identifying valid users.

---

## 🔄 Authentication Flow


Client Request
↓
Extract IP Address
↓
Check IP Rate Limit (Redis)
↓
Check Account Lock (Redis)
↓
Fetch User (PostgreSQL)
↓
Compare Password (bcrypt)
↓
If Failed:
→ Increment Attempts (Redis)
→ Apply Delay
→ Lock Account if Threshold Reached
↓
If Success:
→ Reset Attempts
→ Allow Access


---

## 🧠 Design Decisions

### Why Redis?
- Fast atomic operations (`INCR`)
- Built-in TTL (automatic expiration)
- Scales across multiple servers

---

### Why not store attempts in DB?
- High-frequency writes are inefficient
- Increases database load unnecessarily

---

### Why progressive delay?
- Slows down attackers
- Maintains usability for legitimate users

---

## ⚠️ Security Best Practices Implemented

- ✅ Constant-time password comparison  
- ✅ Generic error messages  
- ✅ Rate limiting (IP + user)  
- ✅ Temporary account lock (not permanent)  
- ✅ No sensitive data exposure  

---

## 🧪 Testing

Try the following:

- Enter wrong password multiple times → account locks  
- Send many requests from same IP → rate limiting triggers  
- Restart server → Redis still enforces limits  

---

## 🚀 Future Improvements

- 🔑 JWT authentication (access + refresh tokens)
- 📱 Multi-factor authentication (MFA)
- 🛡️ CAPTCHA after repeated failures
- 🌍 Geo-location anomaly detection
- 📊 Monitoring & alert system

---

## 🤝 Contributing

Contributions are welcome!  
Feel free to open issues or submit pull requests.

---

## 📜 License

MIT License

---

## 💡 Author Notes

This project demonstrates real-world backend security techniques such as:

- Brute-force attack prevention  
- Timing attack mitigation  
- Distributed rate limiting  

Designed with scalability and security in mind.

