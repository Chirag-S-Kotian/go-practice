// Add JavaScript for interactivity here
// Animated background for hero section
window.addEventListener('DOMContentLoaded', () => {
    // Reveal sections on scroll
    const reveals = document.querySelectorAll('.reveal');
    const revealOnScroll = () => {
        for (const el of reveals) {
            const windowHeight = window.innerHeight;
            const elementTop = el.getBoundingClientRect().top;
            if (elementTop < windowHeight - 80) {
                el.classList.add('visible');
            }
        }
    };
    window.addEventListener('scroll', revealOnScroll);
    revealOnScroll();

    // Animated hero canvas (bubbles)
    const canvas = document.getElementById('hero-canvas');
    if (canvas) {
        const ctx = canvas.getContext('2d');
        let width = canvas.width = window.innerWidth;
        let height = canvas.height = document.querySelector('.hero-bg').offsetHeight;
        window.addEventListener('resize', () => {
            width = canvas.width = window.innerWidth;
            height = canvas.height = document.querySelector('.hero-bg').offsetHeight;
        });
        const bubbles = Array.from({length: 32}, () => ({
            x: Math.random() * width,
            y: Math.random() * height,
            r: 8 + Math.random() * 18,
            d: 1 + Math.random() * 2,
            alpha: 0.2 + Math.random() * 0.5
        }));
        function draw() {
            ctx.clearRect(0, 0, width, height);
            for (const b of bubbles) {
                ctx.beginPath();
                ctx.arc(b.x, b.y, b.r, 0, 2 * Math.PI);
                ctx.fillStyle = `rgba(0,230,255,${b.alpha})`;
                ctx.fill();
                b.y -= b.d;
                if (b.y + b.r < 0) {
                    b.x = Math.random() * width;
                    b.y = height + b.r;
                }
            }
            requestAnimationFrame(draw);
        }
        draw();
    }
});
