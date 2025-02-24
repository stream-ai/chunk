#!/usr/bin/env ts-node

import * as fs from "fs";
import * as ts from "typescript";

interface Chunk {
  file: string;
  chunk_id: string;
  chunk_type: string;
  start_line: number;
  end_line: number;
  code: string;
}

// ...the rest of your script...


interface Chunk {
  file: string;
  chunk_id: string;
  chunk_type: string;
  start_line: number;
  end_line: number;
  code: string;
}

async function main() {
  const filePath = process.argv[2]; // e.g. "myfile.ts" from CLI
  if (!filePath) {
    console.error("Usage: tsparser.ts <filepath>");
    process.exit(1);
  }

  const sourceText = fs.readFileSync(filePath, "utf8");
  const sourceFile = ts.createSourceFile(
    filePath,
    sourceText,
    ts.ScriptTarget.ESNext,
    true,
    filePath.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS
  );

  const chunks: Chunk[] = [];

  function visit(node: ts.Node) {
    // Example: detect function declarations
    if (ts.isFunctionDeclaration(node) && node.name) {
      chunks.push(makeChunk(node, sourceFile, sourceText, filePath, "function_declaration"));
    }
    // Additional logic for arrow functions, class components, etc.

    node.forEachChild(visit);
  }
  visit(sourceFile);

  console.log(JSON.stringify(chunks, null, 2));
}

function makeChunk(
  node: ts.Node,
  sourceFile: ts.SourceFile,
  sourceText: string,
  filePath: string,
  chunkType: string
): Chunk {
  const start = sourceFile.getLineAndCharacterOfPosition(node.getStart());
  const end = sourceFile.getLineAndCharacterOfPosition(node.getEnd());

  return {
    file: filePath,
    chunk_id: chunkType,
    chunk_type: chunkType,
    start_line: start.line + 1,
    end_line: end.line + 1,
    code: sourceText.slice(node.getStart(), node.getEnd())
  };
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
