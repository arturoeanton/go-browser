package main

import (
	"fmt"

	realdom "go-browser/dom"
	"go-browser/spidergopher"
)

func main() {
	fmt.Println("🕷️ Simple DOM Bridge Test (no async)")
	fmt.Println("==================================================")

	// Create engine
	engine := spidergopher.NewEngine()
	// NOTE: NOT calling engine.Start() to avoid async complexity

	// Create DOM tree
	html := realdom.NewElement("html")
	body := realdom.NewElement("body")
	html.AppendChild(body)

	div := realdom.NewElement("div")
	div.Attributes = map[string]string{"id": "test-div", "class": "container"}
	body.AppendChild(div)

	p := realdom.NewElement("p")
	p.AppendChild(realdom.NewText("Hello from DOM!"))
	div.AppendChild(p)

	// Connect to real DOM
	engine.SetDOM(html)

	// Run synchronous script
	_, err := engine.Run(`
		console.log("=== DOM Bridge Test ===");
		
		var el = document.getElementById("test-div");
		console.log("getElementById result:", el ? el.tagName : "null");
		console.log("className:", el ? el.className : "null");
		
		var p = document.querySelector("p");
		console.log("querySelector p:", p ? p.tagName : "null");
		console.log("p.textContent:", p ? p.textContent : "null");
		
		console.log("body exists:", document.body ? "yes" : "no");
		
		console.log("");
		console.log("✅ DOM Bridge working!");
	`)

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("==================================================")
	fmt.Println("👋 Done!")
}
