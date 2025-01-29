# OnDemandCompute

**OnDemandCompute** is a project I started after gaining exposure to the amazing Go programming language. While it might not be perfect, it reflects my efforts to apply the knowledge I’ve gained from contributing to cloud-native projects.

OnDemandCompute is inspired by services like AWS EC2 and other cloud providers offering compute units allocated on demand. This project replicates similar functionality, with features including:

- **Resource Allocation**: Dynamic allocation of resources based on Docker configurations.
- **Exposed SSH Ports**: Simplified access to compute instances.
- **Resource Managers**: Efficiently manage and ensure the availability of resources on the system.

The system’s functionality is exposed through both **gRPC** and **REST APIs** using the gRPC Gateway, making it versatile and accessible.

This project serves as a learning platform and a stepping stone in understanding and replicating real-world cloud computing services.

---
 
## Running the project locally 

### Clone the repository 
 - Clone the repsository to your local system using the command 
 ```git clone https://github.com/chahatsagarmain/OnDemandCompute.git``` 
- Use makefile commands to locally run the project   
```make -f Makefile run```
- The logger should should start logging in the terminal 
- - gRPC api can also be used but its much easier to access using REST api
- A instance of variable resource allocation can be created by simply sending a post request to
``` http://localhost:8080/resource ```
with a JSON Request body as shown . This starts a instance on port 2225 . 

  

      {
        "sshPort" : "2225"
       }
 
# OnDemandCompute REST API Documentation

## Base URL

```
http://localhost:8080
```

---

## **Allocate a Resource**

### **POST** `/v1/resource`

Allocates a new compute resource.

### **Request Body (JSON)**

```json
{
  "mem_required": 1024,
  "disk_required": 20480,
  "cpu_required": 2,
  "ssh_port": "2225"
}
```

### **Response Body (JSON)**

```json
{
  "done": true,
  "message": "Resource allocated successfully"
}
```

---

## **Delete an Allocated Resource**

### **DELETE** `/v1/resource/{id}`

Deletes an allocated compute resource.

### **Path Parameter**

- `id` (string) - The container ID of the resource to delete.

### **Response Body (JSON)**

```json
{
  "done": true,
  "message": "Resource deleted successfully"
}
```

---

## **Get Allocated Resources**

### **GET** `/v1/resource`

Retrieves a list of currently allocated compute resources.

### **Response Body (JSON)**

```json
{
  "Containers": [
    {
      "containerId": "123456789",
      "state": "running",
      "status": "healthy",
      "image": "ubuntu:latest",
      "imageId": "sha256:abcd1234",
      "ports": ["2225"]
    }
  ]
}
```

---

## **Get Container Statistics**

### **GET** `/v1/resource/{id}`

Fetches statistics of a specific container.

### **Path Parameter**

- `id` (string) - The container ID.

### **Response Body (JSON)**

```json
{
  "containerStats": "CPU: 10%, Memory: 512MB"
}
```

---

