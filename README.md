TikTok Shop Web Application:

Overview:

    This project is a simple marketplace web application where users can submit products, search for products, and receive recommendations based on tags. The application uses a combination of technologies including Go, MySQL, Redis, Docker, and WebSockets to provide real-time notifications and efficient data management.

Features:

    Product Submission: Users can submit new products by providing a name, tags, and an optional image.

    Search Functionality: Users can search for products based on name or tags.

    Recommendations: The system provides product recommendations based on the tags of the searched products.

    Real-Time Notifications: Notifications are pushed to connected users in real-time when a new product is submitted.

    WebSocket Integration: Live updates and notifications using WebSockets.

    Docker Integration: The application can be containerized and deployed using Docker.


Technologies Used:

    Go: The primary programming language for the backend.

    MySQL: Relational database management system used to store product data.

    Redis: In-memory data structure store used for caching and real-time notifications.

    Docker: Used to containerize the application for easy deployment.

    WebSockets: Enables real-time communication between the server and clients.


Getting Started - 

Prerequisites

    - Go installed on your machine

    - MySQL server running

    - Redis server running

    - Docker installed

    - Setup

Clone the repository:
    - bash

    - git clone https://github.com/your-username/tiktokshop-webapp.git

    - cd tiktokshop-webapp

Setup MySQL Database:

    - Create a MySQL database named TikTok_Hackathon.

    - Create a table with the following structure:
    
        CREATE TABLE TikTok_Shop (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            tags VARCHAR(255) NOT NULL,
            image_url TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

Setup Redis:

    - Ensure Redis server is running on localhost:6379.

Build and Run the Application:

    - If you're using Docker:

       - bash

       - docker build -t tiktokshop-app .

       - docker run -p 8080:8080 tiktokshop-app

    - Without Docker:

       - bash

       - go run main.go

Access the Application:

    - Open your browser and navigate to http://localhost:8080.

Usage
    
    - Submitting a Product
       
       - Navigate to the homepage.
       
       - Fill in the product name, tags, and optionally upload an image.
       
       - Click "Submit" to add the product.
    
    - Searching for Products
       
       - Enter a search query in the search bar and click "Search".
       
       - Relevant products will be displayed along with recommendations.
    
    - Receiving Real-Time Notifications
       
       - When a new product is added, users connected to the site will receive a real-time notification.

Contributing
    
    - Fork the repository.
    
    - Create your feature branch (git checkout -b feature/AmazingFeature).
        
        - Commit your changes (git commit -m 'Add some AmazingFeature').
        
        - Push to the branch (git push origin feature/AmazingFeature).
        
        - Open a Pull Request.

