
void setup(){
  size(1000,1000);
  loadInfo();
  noiseDetail(12,0.65);
  loadTemps();
  generateNoise();
  print("done");
  drawText();
  drawDays();
  drawName();
  save("output.png");
  exit();
}

String name = "";
float lat = 0;
float lon = 0;

float[] temps = new float[0];
float[] tempDates = new float[0];
float[] daysInMonth = {
  31,
  28,
  31,
  30,
  31,
  30,
  31,
  31,
  30,
  31,
  30,
  31,
};


String[] months = {
    "JAN",
    "FEB",
    "MAR",
    "APR",
    "MAY",
    "JUN",
    "JUL",
    "AUG",
    "SEP",
    "OCT",
    "NOV",
    "DEC",
};


color[] monthColors = {
  color(204, 220, 245),
  color(204, 220, 245),
  color(230, 245, 224),
  color(230, 245, 224),
  color(230, 245, 224),
  color(241, 214, 214),
  color(241, 214, 214),
  color(241, 214, 214),
  color(245, 230, 203),
  color(245, 230, 203),
  color(245, 230, 203),
  color(204, 220, 245),
  
};
void generateNoise(){
  // Before we deal with pixels
  loadPixels();
  // Loop through every pixel
  for (int i = 0; i < pixels.length; i++) {
    float y = floor(i/1000);
    float x = i - y*1000;
    
    float dist = sqrt(pow(y-500,2) + pow(x-500,2));
    if(dist > 500 || dist < 200){
      pixels[i] = color(0);
      continue;
    }
    float val = pow( 1-abs(noise(x/400,y/400) - noise(x/100 +100,y/100 + 150)),5);
    
    
    val = max(val,0.2);
    val *= pow((dist-200)/300,3);
    if(dist > 475){
      if(dist > 490){
        val = 1;
      }else{
        val = max(pow((dist-475)/15,0.7),val);
      }
      
    }
    
    float sin = (y-500)/dist;
    float cos = (x-500)/dist;
    float ang = (atan2(-sin,cos))/(PI*2);
    ang -= 0.25;
    if(ang < 1) ang += 1;
    if(ang > 1) ang -= 1;
    
    ang = 1-ang;
    float temp = interpolateValue(tempDates,temps,ang);
    
    pixels[i] = color(val*255*temp,val*255*pow(1-abs(temp-0.5)*2,2),val*255*(1-temp));
    
  }
  // When we are finished dealing with pixels
  updatePixels();
}


float minTemp = 243.15;
float maxTemp = 313.15;

void loadTemps(){
  Table table = loadTable("calendar-MOD11A2-061-results.csv", "header");
  for (TableRow row : table.rows()) {

    float temp = row.getFloat("MOD11A2_061_LST_Day_1km");
    temp = (constrain(temp,minTemp,maxTemp)-minTemp)/(maxTemp-minTemp);
    String date = row.getString("Date");
    float dateVal = dateToFloat(date);
    if(temp != 0 && dateVal >= 0){
     temps = append(temps,temp);
     tempDates = append(tempDates,dateVal); 
    }
    
  }
  println(temps.length);
}

float dateToFloat(String date){
  String[] parts = split(date,'-');
  int month = Integer.parseInt(parts[1]);
  int day = Integer.parseInt(parts[2]);
  int year = Integer.parseInt(parts[0]);
  float dayPart = (day-1)/daysInMonth[month-1];
  float val = (month+dayPart)/13.0;
  if(year == 2021){
     val -= 1 ;
  }
  return val;
}

float interpolateValue(float[] positions, float[]values, float pos){
 // println(pos);
   int upper = 0;
   if(pos < 0){
     pos += 1;
   }else if(pos > 1){
     pos -= 1;
   }
  boolean wrapUp = true;
 for(int step=0;step< positions.length;step++){
   
   if(positions[step] >= pos){
     upper = step;
     wrapUp = false;
     break;
   }
 }
 


 float maxVal = values[upper];
 int lower = upper-1;

 boolean wrapDown = false;
 if(lower < 0){
   lower += positions.length;
   wrapDown = true;
 }
 float minVal = values[lower];
 
  float maxPos = positions[upper] + (wrapUp ? 1 : 0);;
 float minPos = positions[lower] - (wrapDown&&!wrapUp ? 1 : 0);
 //if(wrapDown) println(minPos + " | " + maxPos + " | " + pos);
 
 float valRange = maxVal-minVal;
 float posRange =abs(maxPos-minPos);
 
 
 float distance = abs(pos-minPos)/posRange;
 return minVal + valRange*distance;

}

void drawText(){
  PFont font = createFont("data/Venera-900.otf", 32);
  textFont(font);
 
  
  int firstDay = 0;
  int dayIndex = 0;
  
  for(int i=0; i<months.length;i++){
    fill(monthColors[i]);
    //float ang = (float(i)/months.length)*PI*2 - PI/2;
    float ang = (firstDay/365.0)*PI*2 - PI/2;
    firstDay += daysInMonth[dayIndex];
    dayIndex++;
    if(ang < 0) ang += PI*2;
    if(ang > PI*2) ang -= PI*2;
    float x = cos(ang)*425;
    float y = sin(ang)*425;
    
    pushMatrix();
    translate(500+x, 500+y);
    //textSize(32);
    float textAng = ang;
    if(textAng > PI*2 ) textAng -= PI*2;
    if(textAng < PI) textAng += PI;
    textAng += PI/2;
    rotate(textAng);
    textAlign(CENTER);
    text(months[i], 0, 0);
    //line(0, 0, 150, 0);
    popMatrix(); 
  }
}

void drawDays(){
      float target = 0;
      int targetId = -1;
      fill(200);
      stroke(50);
      strokeWeight(1);
      for(int i=0; i<365; i++){
        float size = 4;
        if(i == target){
           targetId += 1;
           target += daysInMonth[targetId];
           size = 8;
           fill(255);
        }
        float ang = (float(i)/365)*PI*2 - PI/2;
        if(ang < 0) ang += PI*2;
        if(ang > PI*2) ang -= PI*2;
        float x = cos(ang)*475;
        float y = sin(ang)*475;
        
        circle(x+500,y+500,size);
      }
}

void drawName(){
  textSize(64);
  textAlign(CENTER);
  fill(255);
  text(name,500,480);
  textSize(24);
  text(nf(lat,2,3) + " • " + nf(lon,3,3), 500, 535);
  stroke(170);
  strokeWeight(3);
  line(425,500,575,500);
}

void loadInfo(){
  String[] parts = loadStrings("data.txt");
  name = parts[0];
  lat = float(parts[1]);
  lon = float(parts[2]);
}
