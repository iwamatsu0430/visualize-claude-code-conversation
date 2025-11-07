function toggleThinking(button) {
  const thinkingContent = button.nextElementSibling;
  if (thinkingContent && thinkingContent.classList.contains('thinking-content')) {
    if (thinkingContent.style.display === 'none') {
      thinkingContent.style.display = 'block';
      button.textContent = '🧠 ×';
    } else {
      thinkingContent.style.display = 'none';
      button.textContent = '🧠 ...';
    }
  }
}

function toggleSessionSummary(button) {
  const summaryContent = button.parentElement.nextElementSibling;
  if (summaryContent && summaryContent.classList.contains('session-continuation-content')) {
    if (summaryContent.style.display === 'none') {
      summaryContent.style.display = 'block';
      button.textContent = '📋 Hide conversation summary';
    } else {
      summaryContent.style.display = 'none';
      button.textContent = '📋 View conversation summary';
    }
  }
}

function toggleToolDetails(event) {
  event.stopPropagation();

  const toolItem = event.currentTarget;
  const toolDetails = toolItem.querySelector('.tool-details');

  if (toolDetails) {
    // Close all other open tool details
    document.querySelectorAll('.tool-details').forEach(details => {
      if (details !== toolDetails && details.style.display === 'block') {
        details.style.display = 'none';
      }
    });

    // Toggle current tool details
    if (toolDetails.style.display === 'none') {
      toolDetails.style.display = 'block';
    } else {
      toolDetails.style.display = 'none';
    }
  }
}


// Scroll to message
function scrollToMessage(messageId) {
  const element = document.getElementById(messageId);
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' });
    // Highlight animation
    element.style.backgroundColor = '#fffacd';
    setTimeout(() => {
      element.style.backgroundColor = '';
      element.style.transition = 'background-color 1s ease';
    }, 100);
    setTimeout(() => {
      element.style.transition = '';
    }, 1100);

    // Update active TOC item after scrolling
    setTimeout(() => {
      updateActiveTocItem();
    }, 500);
  }
}

// Jump to previous/next message
function jumpToMessage(button, direction) {
  const currentMessage = button.closest('.user-message');
  const currentId = currentMessage.id;
  const match = currentId.match(/user-msg-(\d+)/);

  if (match) {
    const currentIndex = parseInt(match[1]);
    const targetIndex = direction === 'prev' ? currentIndex - 1 : currentIndex + 1;
    const targetId = `user-msg-${targetIndex}`;
    scrollToMessage(targetId);
  }
}

// Scroll to top
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

// Update active TOC item based on scroll position
function updateActiveTocItem() {
  const userMessages = document.querySelectorAll('.user-message');
  const tocItems = document.querySelectorAll('.toc-item');

  let activeIndex = -1;
  const scrollPosition = window.scrollY + 100; // Offset for better UX

  // Find the current message in view
  for (let i = userMessages.length - 1; i >= 0; i--) {
    const message = userMessages[i];
    if (message.offsetTop <= scrollPosition) {
      // Extract message index from ID (user-msg-0, user-msg-1, etc.)
      const match = message.id.match(/user-msg-(\d+)/);
      if (match) {
        activeIndex = parseInt(match[1]);
      }
      break;
    }
  }

  // Update TOC highlighting
  tocItems.forEach((item, index) => {
    if (index === activeIndex) {
      item.classList.add('active');
    } else {
      item.classList.remove('active');
    }
  });
}

// Show/hide scroll to top button and update active TOC item
let scrollTimeout;
window.addEventListener('scroll', function() {
  const scrollTopBtn = document.getElementById('scrollTopBtn');
  if (window.scrollY > 300) {
    scrollTopBtn.classList.add('show');
  } else {
    scrollTopBtn.classList.remove('show');
  }

  // Debounce the TOC update for better performance
  clearTimeout(scrollTimeout);
  scrollTimeout = setTimeout(updateActiveTocItem, 50);
});

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
  // Set initial active TOC item
  updateActiveTocItem();
});
