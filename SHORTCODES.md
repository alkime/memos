# Hugo Shortcodes

This document describes the custom Hugo shortcodes available in `layouts/_shortcodes/`.

## byline

Displays post author and date information, with optional voice transcription note.

**Usage:**

```markdown
{{< byline >}}
```

**Parameters:**

None - the shortcode pulls data from the page's front matter:
- `author` - from page params or site config
- `date` - from page date
- `voiceBased` - boolean flag in front matter

**Example front matter:**

```yaml
---
title: "My Post"
date: 2025-01-15
author: "John Doe"
voiceBased: true
---
```

**Output:**

> -- posted by **John Doe** on **15 Jan, 2025**
>
> _This post was transcribed from a voice memo by the author and later edited for clarity by the author with help from AI._

---

## image-caption

Displays an image with an optional right-aligned caption below it.

**Usage:**

```markdown
{{< image-caption src="/images/photo.jpg" alt="Description of image" caption="This is the caption text" >}}
```

**Parameters:**

- `src` (required) - path to the image file
- `alt` (optional) - alt text for accessibility
- `caption` (optional) - caption text displayed right-aligned below the image

**Example:**

```markdown
{{< image-caption src="/static/photos/sunset.jpg" alt="Beautiful sunset over the ocean" caption="Taken at Malibu Beach, 2025" >}}
```

**Output:**

- Image displayed at full width
- Caption appears right-aligned below the image in italic gray text
- If no caption is provided, only the image is rendered
