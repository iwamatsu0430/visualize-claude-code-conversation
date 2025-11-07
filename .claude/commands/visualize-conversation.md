---
description: Visualize the current conversation as HTML
allowed-tools: [Bash]
---

Generate an interactive HTML visualization of the current conversation:

!OUTPUT_DIR="${VISUALIZE_OUTPUT_DIR:-./dist}" && ./visualize-conversation "$OUTPUT_DIR" && echo "✅ Visualization created: $OUTPUT_DIR/index.html" && echo "📂 Open: file://$(pwd)/$OUTPUT_DIR/index.html"
