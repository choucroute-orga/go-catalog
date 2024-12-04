// mongo-init.js
const dbName = process.env.MONGODB_DATABASE;
const username = process.env.MONGODB_USERNAME;
const password = process.env.MONGODB_PASSWORD;
const ingredientsCollectionName = process.env.MONGODB_INGREDIENTS_COLLECTION;

db = db.getSiblingDB(dbName);

// Create user with appropriate permissions
db.createUser({
    user: username,
    pwd: password,
    roles: [
        {
            role: 'readWrite',
            db: dbName
        }
    ]
});

// Create ingredients collection
db.createCollection(ingredientsCollectionName);

// Load and insert data from ingredients.json
const ingredientsData = JSON.parse(cat('/docker-entrypoint-initdb.d/ingredients.json'));

// Insert all ingredients
db[ingredientsCollectionName].insertMany(ingredientsData);

// Create indexes if needed (example)
//db.ingredients.createIndex({ "name": 1 });

// Verify the data was inserted
print(`Inserted ${db[ingredientsCollectionName].count()} ingredients insert in collection ${ingredientsCollectionName}into ${dbName} database`);