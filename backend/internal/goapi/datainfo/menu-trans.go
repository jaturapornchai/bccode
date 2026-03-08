package datainfo

import (
	"smlcloudplatform/internal/goapi/logger"
	"math/rand"
	"time"
)

func TransactionInfoHtml(shopId string, mode int) string {
	// Random seed based on current time
	rand.Seed(time.Now().UnixNano())

	// Array of advertisements
	ads := []map[string]string{
		{
			"subtitle": "AI เปลี่ยนธุรกิจ SMEs ให้เติบโตอย่างยั่งยืน",
			"color":    "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
			"features": `
				<h3 style="color: #667eea; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">💡 ทำไมต้อง AI?</h3>
				<p style="margin-bottom: 20px; color: #34495e; line-height: 1.9;">ในยุคดิจิทัล การนำ AI มาใช้ในธุรกิจไม่ใช่เรื่องไกลตัวอีกต่อไป ระบบ AI สามารถช่วยวิเคราะห์พฤติกรรมลูกค้า ทำนายยอดขาย และตอบคำถามลูกค้าแบบอัตโนมัติตลอด 24 ชั่วโมง</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #667eea;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">✨ ประโยชน์ที่ได้รับ:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">มีเวลามากขึ้นในการวางแผนกลยุทธ์ธุรกิจ</li>
						<li style="margin-bottom: 10px;">พัฒนาสินค้าและบริการให้ตรงใจลูกค้า</li>
						<li style="margin-bottom: 10px;">สร้างความสัมพันธ์ที่ดีกับลูกค้า</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #667eea; text-align: center; font-size: 1.15em; padding: 15px; background: #f0f4ff; border-radius: 8px;">🚀 การลงทุนใน AI วันนี้ คือการลงทุนในอนาคตของธุรกิจคุณ</p>
			`,
			"bg": "#667eea",
		},
		{
			"subtitle": "Content Marketing ยุคใหม่ ด้วยพลัง AI",
			"color":    "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
			"features": `
				<h3 style="color: #f5576c; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">📱 คอนเทนต์คือกุญแจความสำเร็จ</h3>
				<p style="margin-bottom: 20px; color: #34495e; line-height: 1.9;">คอนเทนต์คือกุญแจสำคัญในการสร้างแบรนด์ในโลกออนไลน์ แต่การสร้างคอนเทนต์คุณภาพอย่างต่อเนื่องต้องใช้เวลาและความคิดสร้างสรรค์มาก</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #f5576c;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">🎯 AI ช่วยอะไรได้บ้าง:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">✍️ เขียนโฆษณาที่น่าสนใจและเข้าถึงกลุ่มเป้าหมาย</li>
						<li style="margin-bottom: 10px;">🎨 สร้างภาพประกอบที่สวยงามและโดดเด่น</li>
						<li style="margin-bottom: 10px;">📊 วางแผนโพสต์ Social Media อย่างมีประสิทธิภาพ</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #f5576c; text-align: center; font-size: 1.15em; padding: 15px; background: #fff5f7; border-radius: 8px;">💰 ผลลัพธ์: เข้าถึงลูกค้าได้มากขึ้น ยอดขายเพิ่มขึ้นอย่างต่อเนื่อง</p>
			`,
			"bg": "#f5576c",
		},
		{
			"subtitle": "การจัดการสต็อกอัจฉริยะ ลดต้นทุน เพิ่มกำไร",
			"color":    "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
			"features": `
				<h3 style="color: #00b8d4; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">📦 ปัญหาสต็อกสินค้า</h3>
				<p style="margin-bottom: 20px; color: #34495e; line-height: 1.9;">สินค้าคงคลังคือหนึ่งในต้นทุนที่ใหญ่ที่สุดของธุรกิจ การมีสินค้ามากเกินไปทำให้เงินทุนหมุนเวียนช้า ส่วนสินค้าขาดก็ทำให้พลาดโอกาสขาย</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #00b8d4;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">🤖 AI ช่วยได้อย่างไร:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">📊 วิเคราะห์ข้อมูลการขายในอดีต</li>
						<li style="margin-bottom: 10px;">🔮 ทำนายความต้องการในอนาคต</li>
						<li style="margin-bottom: 10px;">🔔 แจ้งเตือนเมื่อสินค้าใกล้หมด</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #00b8d4; text-align: center; font-size: 1.15em; padding: 15px; background: #e0f7fa; border-radius: 8px;">✅ ผลลัพธ์: บริหารสต็อกได้อย่างมีประสิทธิภาพ ลดของเสีย เพิ่มผลกำไร</p>
			`,
			"bg": "#00f2fe",
		},
		{
			"subtitle": "สร้างความสัมพันธ์กับลูกค้าให้ยั่งยืนด้วย CRM",
			"color":    "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
			"features": `
				<h3 style="color: #00bfa5; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">👥 ลูกค้าเก่าคือทรัพย์สิน</h3>
				<p style="margin-bottom: 20px; background: #fffbea; padding: 18px; border-radius: 10px; border-left: 5px solid #00bfa5; color: #34495e; line-height: 1.9;">
					<strong style="color: #f57f17;">💡 รู้หรือไม่?</strong> การรักษาลูกค้าเก่า 1 คน ถูกกว่าหาลูกค้าใหม่ถึง 5 เท่า
				</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #00bfa5;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">🎯 ระบบ CRM ช่วยอะไร:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">📝 จดจำข้อมูลลูกค้าทุกคน</li>
						<li style="margin-bottom: 10px;">📊 ติดตามประวัติการซื้อ</li>
						<li style="margin-bottom: 10px;">🎂 ส่งข้อความอวยพรวันเกิด</li>
						<li style="margin-bottom: 10px;">🎁 เสนอโปรโมชั่นที่ตรงใจแต่ละคน</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #00bfa5; text-align: center; font-size: 1.15em; padding: 15px; background: #e0f2f1; border-radius: 8px;">💚 ผลลัพธ์: ลูกค้ากลับมาซื้อซ้ำและแนะนำเพื่อนมาอีกมากมาย</p>
			`,
			"bg": "#38f9d7",
		},
		{
			"subtitle": "ระบบขายหน้าร้านยุคใหม่ รวดเร็ว แม่นยำ",
			"color":    "linear-gradient(135deg, #fa709a 0%, #fee140 100%)",
			"features": `
				<h3 style="color: #d84315; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">💳 POS ยุคใหม่ทำอะไรได้บ้าง</h3>
				<p style="margin-bottom: 20px; color: #34495e; line-height: 1.9;">การขายหน้าร้านที่รวดเร็วและถูกต้องคือกุญแจสำคัญในการสร้างประสบการณ์ที่ดีให้ลูกค้า</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #d84315;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">⚡ ความสามารถของระบบ:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">💰 รับชำระเงิน (QR Code, บัตรเครดิต, เงินสด)</li>
						<li style="margin-bottom: 10px;">📦 จัดการสต็อกแบบ Real-time</li>
						<li style="margin-bottom: 10px;">📊 ออกรายงานยอดขายทันที</li>
						<li style="margin-bottom: 10px;">🎯 วิเคราะห์สินค้าขายดี</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #d84315; text-align: center; font-size: 1.15em; padding: 15px; background: #fff3e0; border-radius: 8px;">🚀 ผลลัพธ์: ธุรกิจทันสมัย พร้อมรับลูกค้ายุคใหม่</p>
			`,
			"bg": "#fee140",
		},
		{
			"subtitle": "โฆษณาออนไลน์อย่างมีประสิทธิภาพ ได้ลูกค้าจริง",
			"color":    "linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)",
			"features": `
				<h3 style="color: #0097a7; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">🎯 โฆษณาที่ได้ผลจริง</h3>
				<p style="margin-bottom: 20px; background: #fffbea; padding: 18px; border-radius: 10px; border-left: 5px solid #0097a7; color: #34495e; line-height: 1.9;">
					<strong style="color: #f57f17;">⚠️ คำเตือน:</strong> โฆษณาออนไลน์ไม่ใช่แค่การจ่ายเงินแล้วหวังให้โชคดี แต่ต้องอาศัยข้อมูลและกลยุทธ์ที่ถูกต้อง
				</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #0097a7;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">📈 กลยุทธ์ที่ถูกต้อง:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">🎯 ยิงโฆษณาให้ถูกกลุ่มเป้าหมาย</li>
						<li style="margin-bottom: 10px;">⏰ เลือกเวลาที่เหมาะสม</li>
						<li style="margin-bottom: 10px;">💰 ใช้งบประมาณอย่างคุ้มค่า</li>
						<li style="margin-bottom: 10px;">🤖 ใช้ AI วิเคราะห์และปรับแคมเปญอัตโนมัติ</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #0097a7; text-align: center; font-size: 1.15em; padding: 15px; background: #e0f7fa; border-radius: 8px;">📊 ผลลัพธ์: ได้ลูกค้าใหม่อย่างต่อเนื่อง ROI สูงสุด</p>
			`,
			"bg": "#fed6e3",
		},
		{
			"subtitle": "เว็บไซต์และ Chatbot ดูแลลูกค้าไม่มีวันหยุด",
			"color":    "linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)",
			"features": `
				<h3 style="color: #c2185b; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">🌐 ร้านค้าออนไลน์ที่เปิด 24/7</h3>
				<p style="margin-bottom: 20px; color: #34495e; line-height: 1.9;">ในยุคที่ลูกค้าค้นหาข้อมูลและสั่งซื้อสินค้าตลอด 24 ชั่วโมง การมีเว็บไซต์ที่สวยงามและ Chatbot ที่ช่วยตอบคำถามอัตโนมัติเป็นสิ่งจำเป็น</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #c2185b;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">💬 Chatbot AI ทำอะไรได้:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">💡 ตอบคำถามพื้นฐานทันที</li>
						<li style="margin-bottom: 10px;">🛍️ แนะนำสินค้าที่เหมาะสม</li>
						<li style="margin-bottom: 10px;">📦 รับออเดอร์เบื้องต้น</li>
						<li style="margin-bottom: 10px;">⏰ ทำงาน 24 ชั่วโมง ไม่มีวันหยุด</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #c2185b; text-align: center; font-size: 1.15em; padding: 15px; background: #fce4ec; border-radius: 8px;">✨ ผลลัพธ์: ไม่พลาดโอกาสขายแม้ในตอนกลางคืน</p>
			`,
			"bg": "#fecfef",
		},
		{
			"subtitle": "ข้อมูลคือทอง BI Dashboard ช่วยตัดสินใจอย่างชาญฉลาด",
			"color":    "linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)",
			"features": `
				<h3 style="color: #e65100; margin-bottom: 20px; font-size: 1.4em; font-weight: 600;">📊 ตัดสินใจด้วยข้อมูล ไม่ใช่ความรู้สึก</h3>
				<p style="margin-bottom: 20px; background: #fffbea; padding: 18px; border-radius: 10px; border-left: 5px solid #e65100; color: #34495e; line-height: 1.9;">
					<strong style="color: #f57f17;">💡 หลักการ:</strong> การตัดสินใจทางธุรกิจที่ดีต้องอาศัยข้อมูลที่ถูกต้องและทันสมัย
				</p>
				
				<div style="background: #f8f9fa; padding: 25px; border-radius: 12px; margin: 25px 0; border-left: 4px solid #e65100;">
					<h4 style="color: #495057; margin-bottom: 18px; font-size: 1.2em; font-weight: 600;">📈 Dashboard ช่วยอะไร:</h4>
					<ul style="line-height: 2.2; margin-left: 20px; color: #495057;">
						<li style="margin-bottom: 10px;">📊 แสดงภาพรวมธุรกิจด้วยกราฟที่เข้าใจง่าย</li>
						<li style="margin-bottom: 10px;">⚡ ติดตามยอดขายแบบ Real-time</li>
						<li style="margin-bottom: 10px;">📈 วิเคราะห์แนวโน้มและพยากรณ์อนาคต</li>
						<li style="margin-bottom: 10px;">💡 ค้นหาโอกาสทางธุรกิจใหม่ๆ</li>
					</ul>
				</div>
				
				<p style="font-weight: 600; color: #e65100; text-align: center; font-size: 1.15em; padding: 15px; background: #fff3e0; border-radius: 8px;">🎯 ผลลัพธ์: ปรับกลยุทธ์ได้เร็วและแม่นยำยิ่งขึ้น</p>
			`,
			"bg": "#fcb69f",
		},
	}

	// Random select advertisement
	selectedAd := ads[rand.Intn(len(ads))]

	htmlContent := `<!DOCTYPE html>
<html lang="th">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Advertisement</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: 'Sarabun', 'Segoe UI', Tahoma, sans-serif;
			background: ` + selectedAd["color"] + `;
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 20px;
			animation: gradientShift 15s ease infinite;
		}
		@keyframes gradientShift {
			0%, 100% { background-position: 0% 50%; }
			50% { background-position: 100% 50%; }
		}
		@keyframes float {
			0%, 100% { transform: translateY(0px); }
			50% { transform: translateY(-20px); }
		}
		@keyframes pulse {
			0%, 100% { transform: scale(1); }
			50% { transform: scale(1.05); }
		}
		@keyframes shake {
			0%, 100% { transform: rotate(0deg); }
			25% { transform: rotate(-5deg); }
			75% { transform: rotate(5deg); }
		}
		.ad-container {
			max-width: 900px;
			width: 100%;
			background: white;
			border-radius: 30px;
			box-shadow: 0 30px 80px rgba(0,0,0,0.3);
			overflow: hidden;
			animation: float 6s ease-in-out infinite;
		}
		.ad-header {
			background: ` + selectedAd["color"] + `;
			color: white;
			padding: 50px 40px;
			text-align: center;
			position: relative;
			overflow: hidden;
		}
		.ad-icon {
			font-size: 5em;
			margin-bottom: 20px;
			animation: shake 3s ease-in-out infinite;
			display: inline-block;
		}
		.ad-title {
			font-size: 3em;
			font-weight: bold;
			margin-bottom: 15px;
			text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
			line-height: 1.2;
		}
		.ad-subtitle {
			font-size: 1.5em;
			opacity: 0.95;
			margin-bottom: 20px;
		}
		.ad-content {
			padding: 50px 40px;
		}
		.features-box {
			background: #fefcf3;
			border-radius: 20px;
			padding: 40px;
			margin-bottom: 30px;
			font-size: 1.15em;
			line-height: 1.9;
			text-align: left;
			border: 2px solid #e8e3d3;
			color: #2c3e50;
			box-shadow: 0 2px 10px rgba(0,0,0,0.05);
		}
		.price-box {
			background: ` + selectedAd["color"] + `;
			color: white;
			border-radius: 20px;
			padding: 30px;
			text-align: center;
			margin-bottom: 30px;
			animation: pulse 2s ease-in-out infinite;
		}
		.price-label {
			font-size: 1.2em;
			opacity: 0.9;
			margin-bottom: 10px;
		}
		.price-amount {
			font-size: 2.5em;
			font-weight: bold;
			text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
		}
		.cta-button {
			background: ` + selectedAd["bg"] + `;
			color: white;
			border: none;
			border-radius: 50px;
			padding: 25px 60px;
			font-size: 1.8em;
			font-weight: bold;
			cursor: pointer;
			width: 100%;
			text-align: center;
			text-transform: uppercase;
			letter-spacing: 2px;
			box-shadow: 0 10px 30px rgba(0,0,0,0.2);
			transition: all 0.3s ease;
			animation: pulse 2s ease-in-out infinite;
		}
		.cta-button:hover {
			transform: translateY(-5px);
			box-shadow: 0 15px 40px rgba(0,0,0,0.3);
		}
		.footer-info {
			text-align: center;
			margin-top: 30px;
			padding-top: 20px;
			border-top: 2px solid #eee;
			color: #666;
			font-size: 0.9em;
		}
		.refresh-note {
			background: #fff3cd;
			border: 2px solid #ffc107;
			border-radius: 10px;
			padding: 15px;
			text-align: center;
			margin-top: 20px;
			font-size: 1em;
			color: #856404;
		}
		@media (max-width: 768px) {
			.ad-title {
				font-size: 2em;
			}
			.ad-subtitle {
				font-size: 1.2em;
			}
			.ad-content {
				padding: 30px 20px;
			}
			.cta-button {
				font-size: 1.3em;
				padding: 20px 40px;
			}
		}
	</style>
</head>
<body>
	<div class="ad-container">
		<div class="ad-header">
			<h1 class="ad-title">` + selectedAd["subtitle"] + `</h1>
		</div>
		
		<div class="ad-content">
			<div class="features-box">
				` + selectedAd["features"] + `
			</div>
		</div>
	</div>
</body>
</html>`

	logger.Debug("Generated random advertisement HTML for shopId: %s", shopId)
	return htmlContent
}
