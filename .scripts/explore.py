#!/usr/bin/env python3
import json
import sys
import networkx as nx

def analyze_chunk_relations(json_file, target_id=None, max_depth=2):
    with open(json_file) as f:
        data = json.load(f)
    
    # Build chunk ID lookup table
    chunks_by_id = {chunk["id"]: chunk for chunk in data["chunks"]}
    
    # Create a graph of chunk relationships
    G = nx.DiGraph()
    
    for chunk in data["chunks"]:
        G.add_node(chunk["id"], **{k: v for k, v in chunk.items() 
                                if k not in ["content", "related_chunks"]})
        
        for related_id in chunk.get("related_chunks", []):
            G.add_edge(chunk["id"], related_id)
    
    if target_id:
        if target_id not in chunks_by_id:
            return {"error": f"Chunk ID {target_id} not found"}
        
        # Get the ego network around the target chunk
        neighbors = set()
        current = {target_id}
        
        for i in range(max_depth):
            next_level = set()
            for node in current:
                if node in G:
                    next_level.update(G.successors(node))
                    next_level.update(G.predecessors(node))
            neighbors.update(current)
            current = next_level - neighbors
            if not current:
                break
        
        # Create subgraph with the target and its neighborhood
        subgraph = G.subgraph(neighbors)
        
        # Extract relevant information
        result = {
            "target_chunk": {k: v for k, v in chunks_by_id[target_id].items() 
                           if k != "related_chunks"},
            "related_chunks": [{k: v for k, v in chunks_by_id[node].items() 
                              if k not in ["content", "related_chunks"]} 
                             for node in subgraph.nodes() if node != target_id],
            "relationship_paths": []
        }
        
        # Find meaningful paths between chunks
        for related in result["related_chunks"]:
            for path in nx.all_shortest_paths(subgraph, target_id, related["id"]):
                result["relationship_paths"].append({
                    "from": chunks_by_id[path[0]]["file_path"],
                    "to": chunks_by_id[path[-1]]["file_path"],
                    "via": [{"id": node, "file_path": chunks_by_id[node]["file_path"]} 
                           for node in path[1:-1]] if len(path) > 2 else "direct"
                })
        
        return result
    else:
        # Analyze overall relationship structure
        return {
            "summary": {
                "total_chunks": len(data["chunks"]),
                "chunks_with_relations": sum(1 for c in data["chunks"] if c.get("related_chunks")),
                "avg_relations_per_chunk": sum(len(c.get("related_chunks", [])) for c in data["chunks"]) / len(data["chunks"]),
                "max_relations": max(len(c.get("related_chunks", [])) for c in data["chunks"]),
                "connected_components": nx.number_weakly_connected_components(G)
            },
            "most_connected": [{
                "id": chunk["id"],
                "file_path": chunk["file_path"],
                "language": chunk["language"],
                "relation_count": len(chunk.get("related_chunks", [])),
                "symbols": chunk.get("symbols", [])
            } for chunk in sorted(data["chunks"], 
                                key=lambda x: len(x.get("related_chunks", [])), 
                                reverse=True)[:10]]
        }

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: ./chunk_relation_finder.py chunks.json [chunk_id]")
        sys.exit(1)
    
    target_chunk = sys.argv[2] if len(sys.argv) > 2 else None
    results = analyze_chunk_relations(sys.argv[1], target_chunk)
    print(json.dumps(results, indent=2))