DROP INDEX IF EXISTS idx_transaction_description_embeddings_embedding_hnsw;
DROP INDEX IF EXISTS idx_transaction_description_embeddings_exact_lookup;
DROP TABLE IF EXISTS transaction_description_embeddings;

DROP INDEX IF EXISTS idx_classification_entry_embedding_sector_hnsw;
DROP INDEX IF EXISTS idx_classification_entry_embedding_u2_hnsw;
DROP INDEX IF EXISTS idx_classification_entry_embedding_u1_hnsw;
DROP INDEX IF EXISTS idx_classification_entry_embedding_model_active;
DROP TABLE IF EXISTS classification_entry_embedding;

DROP INDEX IF EXISTS idx_classification_entry_list_type_active;
DROP TABLE IF EXISTS classification_entry;

DROP INDEX IF EXISTS idx_list_entry_embeddings_embedding_hnsw;
DROP TABLE IF EXISTS list_entry_embeddings;
