#!/usr/bin/env python3
import json
import sys
import statistics
from collections import Counter

def analyze_chunk_sizes(json_file):
    with open(json_file) as f:
        data = json.load(f)
    
    # Extract token counts
    token_counts = [chunk["token_count"] for chunk in data["chunks"]]
    
    # Calculate statistics
    stats = {
        "count": len(token_counts),
        "min": min(token_counts),
        "max": max(token_counts),
        "mean": statistics.mean(token_counts),
        "median": statistics.median(token_counts),
        "stdev": statistics.stdev(token_counts) if len(token_counts) > 1 else 0,
        "percentiles": {
            "25th": sorted(token_counts)[int(len(token_counts) * 0.25)],
            "75th": sorted(token_counts)[int(len(token_counts) * 0.75)],
            "90th": sorted(token_counts)[int(len(token_counts) * 0.9)],
            "95th": sorted(token_counts)[int(len(token_counts) * 0.95)]
        }
    }
    
    # Bucket into size ranges
    ranges = [(0, 50), (51, 100), (101, 200), (201, 500), (501, 1000), (1001, float('inf'))]
    buckets = Counter()
    
    for count in token_counts:
        for low, high in ranges:
            if low <= count <= high:
                buckets[f"{low}-{high if high != float('inf') else '+'}" ] += 1
                break
    
    # Find outliers (chunks with extreme sizes)
    outliers = {
        "small": [c for c in data["chunks"] if c["token_count"] < stats["percentiles"]["25th"] / 2],
        "large": [c for c in data["chunks"] if c["token_count"] > stats["percentiles"]["95th"]]
    }
    
    return {
        "statistics": stats,
        "distribution": dict(buckets),
        "outliers": outliers,
        "recommendations": get_recommendations(stats, outliers)
    }

def get_recommendations(stats, outliers):
    recommendations = []
    
    if stats["max"] > 5 * stats["median"]:
        recommendations.append("Consider implementing a max token limit to split very large chunks")
    
    if len(outliers["small"]) > 0.1 * stats["count"]:
        recommendations.append("Consider merging very small chunks with adjacent chunks")
    
    if stats["stdev"] / stats["mean"] > 0.75:
        recommendations.append("High variability detected - consider implementing length normalization in your embedding pipeline")
    
    by_language = {}
    for chunk in outliers["large"]:
        by_language.setdefault(chunk["language"], []).append(chunk)
    
    for lang, chunks in by_language.items():
        if len(chunks) > 3:
            recommendations.append(f"Implement specialized chunking for {lang} to handle large chunks better")
    
    return recommendations

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: ./chunk_stats.py chunks.json")
        sys.exit(1)
    
    results = analyze_chunk_sizes(sys.argv[1])
    print(json.dumps(results, indent=2))