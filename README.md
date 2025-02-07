# HNSWGo

A Golang implementation of the Hierarchical Navigable Small World (HNSW) algorithm for efficient approximate nearest neighbor (ANN) search.

## Features

- Supports insertion of high-dimensional vector data.
- Efficient approximate nearest neighbor search.
- Layered graph-based approach for fast search queries.

## Installation

```sh
go get github.com/lblclass/hnswgo
```
## Usage

### Creating an HNSW Index
```go
package main

import (
	"fmt"
	"math/rand"
	"github.com/lblclass/hnswgo/models"
	"github.com/lblclass/hnswgo"
)

func main() {
	// Initialize HNSW index
	h := hnsw.NewHNSW(efConstruction: 200, M: 16, maxLayers: 5, nm: 1.0)

	// Insert elements
	for i := 0; i < 1000; i++ {
		element := models.Element{
			ID: i,
			Embeddings: randomVector(128),
		}
		h.Insert(element)
	}

	// Perform KNN search
	query := models.Element{
		ID: 1001,
		Embeddings: randomVector(128),
	}
	neighbors := h.KNNSearch(query, 10)
	fmt.Println("Nearest neighbors:", neighbors)
}

func randomVector(dim int) []float64 {
	vec := make([]float64, dim)
	for i := range vec {
		vec[i] = rand.Float64()
	}
	return vec
}
```
## API Reference

### NewHNSW(efConstruction int, M int, maxLayers int, nm float64) *HNSW

Creates a new HNSW index.
- efConstruction: Size of the candidate list.
- M: Maximum connections per node.
- maxLayers: Maximum number of layers.
- nm: Normalization factor for level generation.

#### Insert(q models.Element)

Inserts a new element into the index.

### KNNSearch(q models.Element, K int) []int

Finds K approximate nearest neighbors of a given element.

## License
MIT License

Copyright (c) 2025 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
#### 1. Legal Compliance
This software shall not be used for any illegal activities, including but not limited to:
- Violation of any applicable local, national, or international laws.
- Unauthorized data scraping, hacking, or network intrusion.
- Development or deployment of malware, ransomware, or other malicious software.
#### 2.	Ethical Use

The software must not be used for:
- Surveillance or tracking without consent.
- Discrimination, harassment, or harm to individuals or groups.
- Any form of fraud, identity theft, or misinformation.
#### 3.	Security & Responsibility
Users are responsible for:
- Ensuring their use of the software does not compromise system security.
- Keeping their data and credentials secure.
- Reporting security vulnerabilities to the maintainers.
#### 4.	No Warranty

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

