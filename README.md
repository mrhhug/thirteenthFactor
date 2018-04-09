# thirteenthFactor
Cloud native application intended to stop running CloudFoundry apps caught in crash loop.

## Create Camp Crystal
1. Create user
	<pre>
	uaac user add Jason_Voorhees -p <credhubstoredpw> --emails 'looksLikeAppInCrashLoop-STOPPED'
	</pre>
2. Grant privileges
	<pre>
	uaac member add cloud_controller.admin Jason_Voorhees
	uaac member add uaa.admin Jason_Voorhees
	uaac member add scim.read Jason_Voorhees
	uaac member add scim.write Jason_Voorhees
	</pre>

### Quick Start Guide
1. Download manifest
2. Download release
3. Configure variables in manifest.yml
4. Run cf push
	<pre>
	cf push
	</pre>

### Quick Build Guide
1. Clone repo
2. Configure variables in manifest.yml
3. Run deploy script
	<pre>
	./deploy.bash
	</pre>
